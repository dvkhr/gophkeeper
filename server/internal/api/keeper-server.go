// Package api реализует gRPC-сервер для GophKeeper.
// Предоставляет методы для:
// - регистрации и входа пользователей,
// - хранения и получения зашифрованных данных,
// - синхронизации данных между клиентом и сервером.
//
// Сервер использует:
// - JWT для аутентификации,
// - refresh-токены для долгосрочной сессии,
// - репозитории для работы с базой данных,
// - middleware для проверки токенов.
package api

import (
	"context"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/auth"
	"github.com/dvkhr/gophkeeper/server/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// KeeperServer реализует gRPC-сервис KeeperService.
// Хранит репозиторий и конфигурацию для доступа к БД, аутентификации и генерации токенов.
type KeeperServer struct {
	pb.UnimplementedKeeperServiceServer
	srv *service.Service
}

// NewKeeperServer создаёт новый экземпляр KeeperServer.
func NewKeeperServer(srv *service.Service) *KeeperServer {
	return &KeeperServer{
		srv: srv,
	}
}

// Register обрабатывает gRPC-запрос на регистрацию нового пользователя.
func (s *KeeperServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {

	logger.Logg.Info("Register request", "login", req.Login)

	resp, err := s.srv.Register(ctx, req.Login, string(req.EncryptedPassword))
	if err != nil {
		return nil, err
	}

	logger.Logg.Info("Register: user create", "user_id", resp.UserId, "login", req.Login)
	return resp, nil
}

// Login обрабатывает gRPC-запрос на вход пользователя.
func (s *KeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	logger.Logg.Info("Login request", "login", req.Login)

	resp, err := s.srv.Login(ctx, req.Login, string(req.EncryptedPassword))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// StoreData сохраняет зашифрованные данные пользователя в системе.
func (s *KeeperServer) StoreData(ctx context.Context, req *pb.StoreDataRequest) (*pb.StatusResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing user ID in context")
	}

	if req.Record == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Record is required")
	}

	logger.Logg.Info("Storing data", "type", req.Record.Type, "user", userID, "id", req.Record.Id)

	if err := s.srv.StoreData(ctx, userID, req.Record); err != nil {
		return nil, err
	}

	return &pb.StatusResponse{
		Success: true,
		Message: "Data stored successfully",
	}, nil
}

// GetData возвращает все неудалённые данные пользователя.
// Проверяет, что пользователь авторизован (userID в контексте).
// Загружает все записи из БД через сервис.
func (s *KeeperServer) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.DataResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing user ID in context")
	}

	logger.Logg.Info("Getting all data", "user", userID)

	records, err := s.srv.GetData(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &pb.DataResponse{
		Records: records,
	}, nil
}

// SyncData синхронизирует клиентские данные с сервером.
// Проверяет, что пользователь авторизован (userID в контексте).
// Сохраняет все записи.
// Возвращает все неудаленные данные с сервера.
func (s *KeeperServer) SyncData(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing user ID in context")
	}

	logger.Logg.Info("Syncing data", "count", len(req.Records), "user", userID)

	records, err := s.srv.SyncData(ctx, userID, req.Records)
	if err != nil {
		return nil, err
	}

	return &pb.SyncResponse{
		Records: records,
	}, nil
}

// DeleteData помечает запись как удалённую.
func (s *KeeperServer) DeleteData(ctx context.Context, req *pb.DeleteDataRequest) (*pb.StatusResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing user ID in context")
	}

	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Record ID is required")
	}

	logger.Logg.Info("Deleting data", "record_id", req.Id, "user", userID)

	if err := s.srv.DeleteData(ctx, userID, req.Id); err != nil {
		return nil, err
	}

	return &pb.StatusResponse{
		Success: true,
		Message: "Data deleted successfully",
	}, nil
}

// Refresh обновляет пару токенов (access и refresh) по старому refresh-токену.
func (s *KeeperServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	resp, err := s.srv.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Logout отзывает refresh_token
func (s *KeeperServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	if err := s.srv.Logout(ctx, req.RefreshToken); err != nil {
		return nil, err
	}

	return &pb.LogoutResponse{Success: true}, nil
}
