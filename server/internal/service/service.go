// Package service реализует бизнес-логику.
package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/auth"
	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Repo repository.Repository
	Cfg  *config.Config
}

func New(repo repository.Repository, cfg *config.Config) *Service {
	return &Service{Repo: repo, Cfg: cfg}
}

// Register регистрирует нового пользователя в системе.
func (s *Service) Register(ctx context.Context, login, password string) (*pb.AuthResponse, error) {
	if login == "" {
		return nil, status.Errorf(codes.InvalidArgument, "login is required")
	}
	if password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	userID, err := s.Repo.CreateUser(login, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, status.Errorf(codes.AlreadyExists, "user with this login already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	refreshToken, err := auth.GenerateRefreshToken(s.Repo, userID, *s.Cfg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	accessToken, err := auth.GenerateToken(*s.Cfg, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userID,
	}, nil
}

// Login аутентифицирует пользователя по логину и паролю.
func (s *Service) Login(ctx context.Context, login, password string) (*pb.AuthResponse, error) {
	if login == "" {
		return nil, status.Errorf(codes.InvalidArgument, "login is required")
	}
	if password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	user, err := s.Repo.GetUserByLogin(login)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user")
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	refreshToken, err := auth.GenerateRefreshToken(s.Repo, user.ID, *s.Cfg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	accessToken, err := auth.GenerateToken(*s.Cfg, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// StoreData сохраняет или обновляет запись пользователя.
func (s *Service) StoreData(ctx context.Context, userID string, record *pb.DataRecord) error {
	if record == nil {
		return status.Errorf(codes.InvalidArgument, "Record is required")
	}
	if record.Id == "" {
		return status.Errorf(codes.InvalidArgument, "Record ID is required")
	}

	if err := s.Repo.SaveData(userID, record); err != nil {
		return status.Errorf(codes.Internal, "failed to save data: %v", err)
	}

	return nil
}

// GetData возвращает все неудалённые записи пользователя.
func (s *Service) GetData(ctx context.Context, userID string) ([]*pb.DataRecord, error) {
	records, err := s.Repo.GetAllData(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve data: %v", err)
	}
	return records, nil
}

// SyncData синхронизирует клиентские данные с сервером.
func (s *Service) SyncData(ctx context.Context, userID string, records []*pb.DataRecord) ([]*pb.DataRecord, error) {
	for _, record := range records {
		if record == nil || record.Id == "" {
			continue
		}
		if err := s.Repo.SaveData(userID, record); err != nil {
			logger.Logg.Error("Failed to sync record", "id", record.Id, "error", err)
		}
	}

	remoteRecords, err := s.Repo.GetAllData(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve remote data: %v", err)
	}

	return remoteRecords, nil
}

// DeleteData помечает запись как удалённую.
func (s *Service) DeleteData(ctx context.Context, userID, recordID string) error {
	if recordID == "" {
		return status.Errorf(codes.InvalidArgument, "record ID is required")
	}

	exists, err := s.Repo.DataExistsForUser(recordID, userID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to verify data ownership")
	}
	if !exists {
		return status.Errorf(codes.NotFound, "Data not found or access denied")
	}

	if err := s.Repo.MarkDataAsDeleted(recordID); err != nil {
		return status.Errorf(codes.Internal, "failed to delete data")
	}

	return nil
}

// Refresh обновляет пару токенов (access и refresh) по старому refresh-токену.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*pb.AuthResponse, error) {
	if refreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	revoked, err := s.Repo.IsRefreshTokenRevoked(refreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check token status")
	}
	if revoked {
		return nil, status.Errorf(codes.Unauthenticated, "token revoked")
	}

	userID, err := s.Repo.GetUserIDByRefreshToken(refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}
		return nil, status.Error(codes.Internal, "failed to get user ID")
	}

	newAccessToken, err := auth.GenerateToken(*s.Cfg, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	newRefreshToken, err := auth.GenerateRefreshToken(s.Repo, userID, *s.Cfg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	if err := s.Repo.RevokeRefreshToken(refreshToken); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke old token")
	}

	return &pb.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		UserId:       userID,
	}, nil
}

// Logout отзывает refresh-токен, завершая сессию пользователя.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	if err := s.Repo.RevokeRefreshToken(refreshToken); err != nil {
		return status.Errorf(codes.Internal, "failed to revoke token")
	}

	return nil
}
