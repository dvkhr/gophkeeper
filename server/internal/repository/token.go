package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var _ TokenRepository = (*PostgresTokenRepository)(nil)

// TokenRepository — интерфейс для работы с токенами в базе данных.
type TokenRepository interface {
	// SaveRefreshToken сохраняет refresh-токен для пользователя.
	// Возвращает ошибку, если сохранение не удалось.
	SaveRefreshToken(token, userID string, expiresAt time.Time) error

	// IsRefreshTokenRevoked проверяет, был ли refresh-токен отозван.
	// Возвращает true, если токен не найден или отозван.
	IsRefreshTokenRevoked(token string) (bool, error)

	// RevokeRefreshToken отмечает refresh-токен как отозванный.
	// Возвращает ошибку, если операция не удалась.
	RevokeRefreshToken(token string) error

	// GetUserIDByRefreshToken находит и возвращает идентификатор пользователя по значению refresh-токена.
	// Возвращает ошибку sql.ErrNoRows, если токен не найден или отозван.
	GetUserIDByRefreshToken(token string) (string, error)
}

// PostgresTokenRepository — реализация TokenRepository для PostgreSQL.
type PostgresTokenRepository struct {
	db *sql.DB
}

// NewTokenRepository создаёт новый экземпляр TokenRepository.
func NewTokenRepository(db *sql.DB) TokenRepository {
	return &PostgresTokenRepository{db: db}
}

// SaveRefreshToken сохраняет refresh-токен в базе данных.
func (r *PostgresTokenRepository) SaveRefreshToken(token, userID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(context.Background(),
		`INSERT INTO refresh_tokens (token, user_id, expires_at, revoked)
         VALUES ($1, $2, $3, false)`,
		token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

// IsRefreshTokenRevoked проверяет, был ли refresh-токен отозван.
func (r *PostgresTokenRepository) IsRefreshTokenRevoked(token string) (bool, error) {
	var revoked bool
	err := r.db.QueryRow(`
        SELECT revoked FROM refresh_tokens 
        WHERE token = $1
    `, token).Scan(&revoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return true, err
	}
	return revoked, nil
}

// RevokeRefreshToken отмечает refresh-токен как отозванный.
func (r *PostgresTokenRepository) RevokeRefreshToken(token string) error {
	_, err := r.db.ExecContext(context.Background(),
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

// GetUserIDByRefreshToken находит и возвращает идентификатор пользователя по значению refresh-токена.
func (r *PostgresTokenRepository) GetUserIDByRefreshToken(token string) (string, error) {
	var userID string
	err := r.db.QueryRow(`
        SELECT user_id FROM refresh_tokens 
        WHERE token = $1 AND revoked = false
    `, token).Scan(&userID)
	return userID, err
}
