package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var _ TokenRepository = (*PostgresTokenRepository)(nil)

type TokenRepository interface {
	SaveRefreshToken(token, userID string, expiresAt time.Time) error
	IsRefreshTokenRevoked(token string) (bool, error)
	RevokeRefreshToken(token string) error
}

type PostgresTokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &PostgresTokenRepository{db: db}
}

func (r *PostgresTokenRepository) SaveRefreshToken(token, userID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(context.Background(),
		`INSERT INTO refresh_tokens (token, user_id, expires_at)
         VALUES ($1, $2, $3)`,
		token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

func (r *PostgresTokenRepository) IsRefreshTokenRevoked(token string) (bool, error) {
	var revoked bool
	err := r.db.QueryRowContext(context.Background(),
		`SELECT revoked FROM refresh_tokens WHERE token = $1`, token).Scan(&revoked)
	if err == sql.ErrNoRows {
		return true, nil // токен не найден → считаем отозванным
	}
	if err != nil {
		return true, fmt.Errorf("failed to check refresh token: %w", err)
	}
	return revoked, nil
}

func (r *PostgresTokenRepository) RevokeRefreshToken(token string) error {
	_, err := r.db.ExecContext(context.Background(),
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}
