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

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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

func (s *KeeperServer) StoreData(ctx context.Context, req *pb.StoreDataRequest) (*pb.StatusResponse, error) {
	logger.Logg.Info("Storing data: %v", req.Record.Type)
	return &pb.StatusResponse{Success: true, Message: "Stored"}, nil
}

func (s *KeeperServer) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.DataResponse, error) {
	logger.Logg.Info("Get data by type: %v", req.Type)
	return &pb.DataResponse{}, nil
}

func (s *KeeperServer) SyncData(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	logger.Logg.Info("Syncing %d records", len(req.Records))
	return &pb.SyncResponse{}, nil
}
