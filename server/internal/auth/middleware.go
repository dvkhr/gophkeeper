// pkg/auth/middleware.go

package auth

import (
	"context"
	"strings"

	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor — gRPC middleware для проверки JWT-токена в заголовках.
func AuthInterceptor(cfg config.Config, repo repository.TokenRepository) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Пропускаем методы, не требующие авторизации
		if info.FullMethod == "/keeper.KeeperService/Login" || info.FullMethod == "/keeper.KeeperService/Register" {
			return handler(ctx, req)
		}

		// Получаем токен из заголовков
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization not provided")
		}

		tokenStr := strings.TrimPrefix(values[0], "Bearer ")
		if tokenStr == "" {
			return nil, status.Errorf(codes.Unauthenticated, "empty token")
		}

		// Проверяем токен
		claims, err := ParseToken(cfg, tokenStr)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Проверяем, отозван ли токен
		revoked, err := repo.IsRefreshTokenRevoked(tokenStr)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to check token status")
		}
		if revoked {
			return nil, status.Errorf(codes.Unauthenticated, "token revoked")
		}

		// Добавляем userID в контекст
		ctx = WithUserID(ctx, claims.UserID)
		return handler(ctx, req)
	}
}
