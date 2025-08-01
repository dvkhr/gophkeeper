// pkg/auth/middleware.go

package auth

import (
	"context"
	"strings"

	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor — gRPC middleware для проверки JWT-токена в заголовках.
// Пропускает методы /keeper.KeeperService/Login и /keeper.KeeperService/Register без проверки.
// Для остальных методов:
// - извлекает Bearer-токен,
// - проверяет его валидность,
// - проверяет, не отозван ли он,
// - добавляет userID в контекст.
func AuthInterceptor(cfg config.Config, repo repository.TokenRepository) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/keeper.KeeperService/Login" ||
			info.FullMethod == "/keeper.KeeperService/Register" ||
			info.FullMethod == "/keeper.KeeperService/Refresh" ||
			info.FullMethod == "/keeper.KeeperService/Logout" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Logg.Warn("Metadata not provided")

			return nil, status.Errorf(codes.Unauthenticated, "metadata not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			logger.Logg.Warn("Authorization header not provided")

			return nil, status.Errorf(codes.Unauthenticated, "authorization not provided")
		}

		tokenStr := strings.TrimPrefix(values[0], "Bearer ")
		if tokenStr == "" {
			logger.Logg.Warn("Empty token")

			return nil, status.Errorf(codes.Unauthenticated, "empty token")
		}

		claims, err := ParseToken(cfg, tokenStr)
		if err != nil {
			logger.Logg.Warn("Invalid token", "error", err)

			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = WithUserID(ctx, claims.UserID)
		logger.Logg.Debug("User ID установлен в контекст", "user_id", claims.UserID)

		return handler(ctx, req)
	}
}
