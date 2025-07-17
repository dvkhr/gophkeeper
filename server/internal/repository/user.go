package repository

import (
	"context"
	"database/sql"
	"fmt"
)

var _ UserRepository = (*PostgresUserRepository)(nil)

// UserRepository — интерфейс для работы с пользователями в базе данных.
type UserRepository interface {
	// CreateUser создаёт нового пользователя с указанным логином и хэшем пароля.
	// Возвращает идентификатор созданного пользователя или ошибку.
	CreateUser(login, passwordHash string) (string, error)

	// GetUserByLogin возвращает пользователя по его логину, если он существует и активен.
	// Возвращает nil, если пользователь не найден.
	GetUserByLogin(login string) (*User, error)
}

// PostgresUserRepository — реализация UserRepository для PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт новый экземпляр UserRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser создаёт нового пользователя в базе данных.
// Возвращает идентификатор пользователя или ошибку.
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

// GetUserByLogin ищет пользователя по логину в базе данных.
// Возвращает *User, если пользователь найден и активен.
// Возвращает nil, если пользователь не найден.
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
