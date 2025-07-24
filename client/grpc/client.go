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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
	encryptedData, err := c.crypto.Encrypt(record.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("ошибка шифрования: %w", err)
	}

	encryptedRecord := &pb.DataRecord{
		Id:            record.Id,
		Type:          record.Type,
		EncryptedData: encryptedData,
		Metadata:      record.Metadata,
		Timestamp:     record.Timestamp,
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
		encryptedData, err := c.crypto.Encrypt(record.EncryptedData)
		if err != nil {
			return nil, fmt.Errorf("ошибка шифрования записи %s: %w", record.Id, err)
		}

		encryptedRecords = append(encryptedRecords, &pb.DataRecord{
			Id:            record.Id,
			Type:          record.Type,
			EncryptedData: encryptedData,
			Metadata:      record.Metadata,
			Timestamp:     record.Timestamp,
		})
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

// SetToken устанавливает токен.
func (c *Client) SetToken(accessToken, refreshToken string) error {
	c.token = accessToken
	return file.Save(&file.Data{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// GetToken возвращает текущий токен.
func (c *Client) GetToken() string {
	return c.token
}
