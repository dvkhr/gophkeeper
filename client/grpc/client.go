package grpc

import (
	"context"
	"fmt"

	"github.com/dvkhr/gophkeeper/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client — gRPC-клиент для GophKeeper.
type Client struct {
	conn    *grpc.ClientConn
	service pb.KeeperServiceClient
	token   string
}

// New создаёт новый gRPC-клиент.
// address — адрес сервера, например "localhost:8080"
func New(address string) (*Client, error) {
	clientConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}

	return &Client{
		conn:    clientConn,
		service: pb.NewKeeperServiceClient(clientConn),
	}, nil
}

// Close закрывает соединение с сервером.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Register регистрирует нового пользователя.
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

// Login выполняет вход пользователя.
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

// StoreData сохраняет запись.
func (c *Client) StoreData(record *pb.DataRecord) (*pb.StatusResponse, error) {
	ctx := c.authContext()
	resp, err := c.service.StoreData(ctx, &pb.StoreDataRequest{Record: record})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetData получает все данные пользователя.
func (c *Client) GetData() (*pb.DataResponse, error) {
	ctx := c.authContext()
	resp, err := c.service.GetData(ctx, &pb.GetDataRequest{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SyncData синхронизирует данные.
func (c *Client) SyncData(records []*pb.DataRecord) (*pb.SyncResponse, error) {
	ctx := c.authContext()
	req := &pb.SyncRequest{Records: records}
	resp, err := c.service.SyncData(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
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

// authContext возвращает контекст с заголовком авторизации.
func (c *Client) authContext() context.Context {
	if c.token == "" {
		return context.Background()
	}

	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+c.token))
}

// SetToken устанавливает токен.
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken возвращает текущий токен.
func (c *Client) GetToken() string {
	return c.token
}
