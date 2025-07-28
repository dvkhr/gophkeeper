// Package grpc предоставляет gRPC-клиент для взаимодействия с сервером GophKeeper.
//
// Клиент отвечает за:
//   - Установление безопасного соединения с сервером.
//   - Шифрование и расшифровку данных с использованием AES-GCM.
//   - Аутентификацию через JWT-токены.
//   - Сохранение сессии (токенов и соли) в локальном хранилище.
//
// Все данные шифруются на клиенте, сервер хранит только зашифрованные данные.
package grpc

import (
	"context"
	"fmt"

	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client — gRPC-клиент для GophKeeper.
type Client struct {
	conn    *grpc.ClientConn
	service pb.KeeperServiceClient
	token   string
	crypto  *crypto.Encryptor
}

// New создаёт новый gRPC-клиент и устанавливает соединение с сервером.
// address — адрес сервера, например "localhost:50051"
func New(address string, encryptionKey []byte) (*Client, error) {
	clientConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}

	encryptor, err := crypto.NewEncryptor(encryptionKey)
	if err != nil {
		clientConn.Close()
		return nil, fmt.Errorf("неверный ключ шифрования: %w", err)
	}

	return &Client{
		conn:    clientConn,
		service: pb.NewKeeperServiceClient(clientConn),
		crypto:  encryptor,
	}, nil
}

// Close закрывает соединение с gRPC-сервером.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Register регистрирует нового пользователя на сервере.
func (c *Client) Register(login string, encryptedPassword []byte) (*pb.AuthResponse, error) {
	req := &pb.RegisterRequest{
		Login:             login,
		EncryptedPassword: encryptedPassword,
	}

	resp, err := c.service.Register(context.Background(), req)
	if err != nil {
		return nil, err
	}

	c.token = resp.AccessToken

	return resp, nil
}

// Login выполняет вход существующего пользователя.
func (c *Client) Login(login string, encryptedPassword []byte) (*pb.AuthResponse, error) {
	req := &pb.LoginRequest{
		Login:             login,
		EncryptedPassword: encryptedPassword,
	}

	resp, err := c.service.Login(context.Background(), req)
	if err != nil {
		return nil, err
	}

	c.token = resp.AccessToken

	return resp, nil
}

// StoreData сохраняет одну запись в хранилище.
func (c *Client) StoreData(record *pb.DataRecord) (*pb.StatusResponse, error) {
	encryptedRecord, err := c.encryptRecord(record)
	if err != nil {
		return nil, err
	}

	ctx := c.authContext()
	resp, err := c.service.StoreData(ctx, &pb.StoreDataRequest{Record: encryptedRecord})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetData запрашивает все неудалённые записи пользователя.
func (c *Client) GetData() (*pb.DataResponse, error) {
	ctx := c.authContext()
	resp, err := c.service.GetData(ctx, &pb.GetDataRequest{})
	if err != nil {
		return nil, err
	}

	for _, record := range resp.Records {
		plaintext, err := c.crypto.Decrypt(record.EncryptedData)
		if err != nil {
			logger.Logg.Warn("ошибка расшифрования записи ", record.Id)
			continue
		}
		record.EncryptedData = plaintext
	}

	return resp, nil
}

// SyncData синхронизирует список записей с сервером.
func (c *Client) SyncData(records []*pb.DataRecord) (*pb.SyncResponse, error) {
	var encryptedRecords []*pb.DataRecord
	for _, record := range records {
		encryptedRecord, err := c.encryptRecord(record)
		if err != nil {
			return nil, err
		}
		if encryptedRecord != nil {
			encryptedRecords = append(encryptedRecords, encryptedRecord)
		}
	}

	ctx := c.authContext()
	req := &pb.SyncRequest{Records: encryptedRecords}
	syncResp, err := c.service.SyncData(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, record := range syncResp.Records {
		plaintext, err := c.crypto.Decrypt(record.EncryptedData)
		if err != nil {
			continue
		}
		record.EncryptedData = plaintext
	}

	return syncResp, nil
}

// DeleteData удаляет запись по ID.
func (c *Client) DeleteData(id string) (*pb.StatusResponse, error) {
	ctx := c.authContext()
	req := &pb.DeleteDataRequest{Id: id}
	resp, err := c.service.DeleteData(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// authContext возвращает контекст с заголовком авторизации (Bearer token).
// Если токен не установлен, возвращается пустой контекст.
func (c *Client) authContext() context.Context {
	if c.token == "" {
		return context.Background()
	}

	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+c.token))
}

// SetToken устанавливает токен, не затрагивая другие данные.
func (c *Client) SetToken(accessToken, refreshToken string) error {
	c.token = accessToken

	session, err := file.Load()
	if err != nil {
		return fmt.Errorf("не удалось загрузить сессию: %w", err)
	}

	session.AccessToken = accessToken
	session.RefreshToken = refreshToken

	return file.Save(session)
}

// GetToken возвращает текущий токен.
func (c *Client) GetToken() string {
	return c.token
}

// Refresh обновляет пару токенов (access и refresh) на основе переданного refresh-токена.
func (c *Client) Refresh() error {
	session, err := file.Load()
	if err != nil || session.RefreshToken == "" {
		return fmt.Errorf("нет refresh_token")
	}

	req := &pb.RefreshRequest{
		RefreshToken: session.RefreshToken,
	}

	resp, err := c.service.Refresh(context.Background(), req)
	if err != nil {
		return err
	}

	c.token = resp.AccessToken
	return c.SetToken(resp.AccessToken, resp.RefreshToken)
}

// DoWithRetry выполняет функцию с повторной попыткой при 401
func (c *Client) DoWithRetry(fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		return err
	}

	logger.Logg.Info("Попытка обновить токен...")
	if refreshErr := c.Refresh(); refreshErr != nil {
		logger.Logg.Error("Не удалось обновить токен", "error", refreshErr)
		return err
	}

	return fn()
}

// Logout отзывает refresh_token на сервере
func (c *Client) Logout(refreshToken string) error {
	req := &pb.LogoutRequest{
		RefreshToken: refreshToken,
	}
	_, err := c.service.Logout(context.Background(), req)
	return err
}

// encryptRecord шифрует данные одной записи
func (c *Client) encryptRecord(record *pb.DataRecord) (*pb.DataRecord, error) {
	if record == nil {
		return nil, nil
	}

	encryptedData, err := c.crypto.Encrypt(record.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("ошибка шифрования записи %s: %w", record.Id, err)
	}

	return &pb.DataRecord{
		Id:            record.Id,
		Type:          record.Type,
		EncryptedData: encryptedData,
		Metadata:      record.Metadata,
		Timestamp:     record.Timestamp,
	}, nil
}
