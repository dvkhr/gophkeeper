// Package db предоставляет функции для подключения к базе данных и применения миграций.
package db

import (
	"database/sql"
	"fmt"

	"github.com/dvkhr/gophkeeper/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect устанавливает соединение с базой данных PostgreSQL по заданной строке подключения (DSN).
//
// Возвращает:
//   - *sql.DB — указатель на объект базы данных.
//   - error — ошибку, если подключение не удалось установить.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Logg.Info("Connected to the database")
	return db, nil
}
