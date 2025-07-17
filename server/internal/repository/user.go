package repository

import (
	"context"
	"database/sql"
	"fmt"
)

var _ UserRepository = (*PostgresUserRepository)(nil)

type UserRepository interface {
	CreateUser(login, passwordHash string) (string, error)
	GetUserByLogin(login string) (*User, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(login, passwordHash string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(context.Background(),
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`,
		login, passwordHash).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

func (r *PostgresUserRepository) GetUserByLogin(login string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(context.Background(),
		`SELECT id, login, password_hash, status, EXTRACT(EPOCH FROM created_at)::int, EXTRACT(EPOCH FROM updated_at)::int 
         FROM users WHERE login = $1 AND status = 'active'`,
		login).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}
	return &u, nil
}
