// pkg/auth/token.go

package auth

import (
	"time"

	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// Claims — структура полезной нагрузки токена
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken — создаёт новый JWT-токен для пользователя
func GenerateToken(cfg config.Config, userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.Auth.JWTTTLHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "GophKeeper",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Auth.JWTSecret))
}

// ParseToken — разбирает строку токена и возвращает полезную нагрущку
func ParseToken(cfg config.Config, tokenStr string) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.Auth.JWTSecret), nil
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}
