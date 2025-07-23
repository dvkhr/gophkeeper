package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenRepository_SaveAndCheckToken(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	tokenRepo := NewTokenRepository(db)

	login := "testuser"
	password := "hashedpass"

	// 1. Создаем пользователя
	userID, err := userRepo.CreateUser(login, password)
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	// 2. Подготавливаем токен
	token := "refresh_token_123"
	expiresAt := time.Now().Add(1 * time.Hour)

	// 3. Сохраняем токен
	err = tokenRepo.SaveRefreshToken(token, userID, expiresAt)
	require.NoError(t, err)

	// 4. Проверяем, что он не отозван
	revoked, err := tokenRepo.IsRefreshTokenRevoked(token)
	require.NoError(t, err)
	assert.False(t, revoked)
}

func TestTokenRepository_RevokeToken(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	tokenRepo := NewTokenRepository(db)

	login := "testuser"
	password := "hashedpass"

	// 1. Создаем пользователя
	userID, err := userRepo.CreateUser(login, password)
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	// 2. Подготавливаем токен
	token := "refresh_token_456"
	expiresAt := time.Now().Add(1 * time.Hour)

	// 3. Сохраняем токен
	err = tokenRepo.SaveRefreshToken(token, userID, expiresAt)
	require.NoError(t, err)

	// 4. Отзываем токен
	err = tokenRepo.RevokeRefreshToken(token)
	require.NoError(t, err)

	// 5. Проверяем, что он отозван
	revoked, err := tokenRepo.IsRefreshTokenRevoked(token)
	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestTokenRepository_CheckNonExistentToken(t *testing.T) {
	db := setupTestDB()
	tokenRepo := NewTokenRepository(db)

	// Проверяем несуществующий токен
	revoked, err := tokenRepo.IsRefreshTokenRevoked("nonexistent_token")
	require.NoError(t, err)
	assert.True(t, revoked) // токен не найден → отозван
}
