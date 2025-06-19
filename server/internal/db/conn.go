package db

import (
	"database/sql"
	"fmt"

	"github.com/dvkhr/gophkeeper/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

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
