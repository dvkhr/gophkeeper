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
	"strings"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/auth"
	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// KeeperServer реализует gRPC-сервис KeeperService.
// Хранит репозиторий и конфигурацию для доступа к БД, аутентификации и генерации токенов.
type KeeperServer struct {
	pb.UnimplementedKeeperServiceServer
	repo repository.Repository
	cfg  *config.Config
}

// NewKeeperServer создаёт новый экземпляр KeeperServer.
func NewKeeperServer(repo repository.Repository, cfg *config.Config) *KeeperServer {
	return &KeeperServer{
		repo: repo,
		cfg:  cfg,
	}
}

// Register обрабатывает запрос на регистрацию нового пользователя.
// Ожидает зашифрованный пароль (EncryptedPassword) от клиента.
// Хэширует пароль и сохраняет пользователя в БД.
// Возвращает пару токенов: access и refresh.
func (s *KeeperServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	logger.Logg.Info("Register request", "login", req.Login)

	hashedPassword, err := auth.HashPassword(string(req.EncryptedPassword))
	if err != nil {
		logger.Logg.Error("Failed to hash password", "error", err)
		return nil, err
	}

	userID, err := s.repo.CreateUser(req.Login, hashedPassword)
	if err != nil {
		logger.Logg.Error("Register failed", "error", err)

		// Проверяем, что это ошибка дубликата
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, status.Errorf(codes.AlreadyExists, "user with this login already exists")
		}

		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	refreshToken, err := auth.GenerateRefreshToken(s.repo, userID, *s.cfg)
	if err != nil {
		logger.Logg.Error("Failed to generate refresh token", "error", err)
		return nil, err
	}

	accessToken, err := auth.GenerateToken(*s.cfg, userID)
	if err != nil {
		logger.Logg.Error("Failed to generate access token", "error", err)
		return nil, err
	}
	logger.Logg.Info("Register: user create", "user_id", userID, "login", req.Login)

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userID,
	}, nil
}

// Login обрабатывает запрос на вход.
// Проверяет логин и зашифрованный пароль.
// Если данные верны — возвращает пару токенов.
func (s *KeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	logger.Logg.Info("Login request", "login", req.Login)

	user, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		logger.Logg.Error("Login failed", "error", err)
		return nil, err
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	if !auth.CheckPasswordHash(string(req.EncryptedPassword), user.PasswordHash) {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credentials")
	}

	refreshToken, err := auth.GenerateRefreshToken(s.repo, user.ID, *s.cfg)
	if err != nil {
		logger.Logg.Error("Failed to generate refresh token", "error", err)
		return nil, err
	}

	accessToken, err := auth.GenerateToken(*s.cfg, user.ID)
	if err != nil {
		logger.Logg.Error("Failed to generate access token", "error", err)
		return nil, err
	}

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// StoreData сохраняет зашифрованные данные пользователя в системе.
// Проверяет, что пользователь авторизован (userID в контексте).
// Проверяет, что запись и её ID не пустые.
// Сохраняет или обновляет данные через репозиторий.
func (s *KeeperServer) StoreData(ctx context.Context, req *pb.StoreDataRequest) (*pb.StatusResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Missing user ID in context")
	}

	if req.Record == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Record is required")
	}

	if req.Record.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Record ID is required")
	}

	logger.Logg.Info("Storing data", "type", req.Record.Type, "user", userID)

	err := s.repo.SaveData(userID, req.Record)
	if err != nil {
		logger.Logg.Error("Failed to store data", "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to save data")
	}

	return &pb.StatusResponse{
		Success: true,
		Message: "Data stored successfully",
	}, nil
}

// GetData возвращает все неудалённые данные пользователя.
// Проверяет, что пользователь авторизован (userID в контексте).
// Загружает все записи из БД через репозиторий.
func (s *KeeperServer) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.DataResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Missing user ID in context")
	}

	logger.Logg.Info("Getting all data", "user", userID)

	records, err := s.repo.GetAllData(userID)
	if err != nil {
		logger.Logg.Error("Failed to get data", "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to retrieve data")
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
		return nil, status.Errorf(codes.Unauthenticated, "Missing user ID in context")
	}

	logger.Logg.Info("Syncing data", "count", len(req.Records), "user", userID)

	// сохраняем
	for _, record := range req.Records {
		if record == nil || record.Id == "" {
			continue // пропускаем невалидные
		}
		if err := s.repo.SaveData(userID, record); err != nil {
			logger.Logg.Error("Failed to sync record", "id", record.Id, "error", err)
		}
	}

	// получаем с сервера
	remoteRecords, err := s.repo.GetAllData(userID)
	if err != nil {
		logger.Logg.Error("Failed to fetch remote data", "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to retrieve remote data")
	}

	return &pb.SyncResponse{
		Records: remoteRecords,
	}, nil
}

// DeleteData помечает запись как удалённую (soft delete).
// Проверяет, что пользователь авторизован.
// Проверяет, что запись существует и принадлежит пользователю.
// Устанавливает флаг "deleted = true" в БД.
func (s *KeeperServer) DeleteData(ctx context.Context, req *pb.DeleteDataRequest) (*pb.StatusResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Missing user ID in context")
	}

	logger.Logg.Info("Deleting data", "record_id", req.Id, "user", userID)

	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Record ID is required")
	}

	// Проверяем, существует ли запись и принадлежит ли пользователю
	exists, err := s.repo.DataExistsForUser(req.Id, userID)
	if err != nil {
		logger.Logg.Error("Failed to check data ownership", "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to verify data ownership")
	}
	if !exists {
		return nil, status.Errorf(codes.NotFound, "Data not found or access denied")
	}

	err = s.repo.MarkDataAsDeleted(req.Id)
	if err != nil {
		logger.Logg.Error("Failed to delete data", "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to delete data")
	}

	return &pb.StatusResponse{
		Success: true,
		Message: "Data deleted successfully",
	}, nil
}
