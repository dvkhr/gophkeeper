// pkg/auth/token.go

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/repository"
)

// GenerateRefreshToken генерирует случайный refresh-токен и сохраняет его в БД
func GenerateRefreshToken(repo repository.TokenRepository, userID string, cfg config.Config) (string, error) {
	token := GenerateRandomString(32)
	expiresAt := time.Now().Add(time.Duration(cfg.Auth.RefreshTokenTTLDays) * 24 * time.Hour)

	err := repo.SaveRefreshToken(token, userID, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

// RevokeRefreshToken отзывает refresh-токен
func RevokeRefreshToken(repo repository.TokenRepository, token string) error {
	return repo.RevokeRefreshToken(token)
}

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
