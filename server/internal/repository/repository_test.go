package repository

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/db"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// setupTestDB — инициализирует тестовое соединение с базой данных.
// Применяет миграции один раз и очищает таблицы перед каждым тестом.
func setupTestDB() *sql.DB {
	dsn := "postgres://postgres:postgres@localhost:5432/gophkeeper_test?sslmode=disable"

	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to DB: %v", err))
	}

	err = dbConn.Ping()
	if err != nil {
		panic(fmt.Sprintf("failed to ping DB: %v", err))
	}

	// Очистка перед каждым запуском
	_, err = dbConn.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		panic(fmt.Sprintf("failed to reset schema: %v", err))
	}

	// Применяем миграции
	err = db.ApplyMigrations(dbConn)
	if err != nil {
		panic(fmt.Sprintf("failed to apply migrations: %v", err))
	}

	return dbConn
}

func TestMain(m *testing.M) {
	logger.Logg = logger.NewTestLogger()

	dbConn := setupTestDB()
	defer dbConn.Close()

	os.Exit(m.Run())
}
