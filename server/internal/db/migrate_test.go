package db_test

import (
	"database/sql"
	"testing"

	"github.com/dvkhr/gophkeeper/server/internal/db"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB — создаёт соединение с тестовой БД
func setupTestDB(t *testing.T) *sql.DB {
	dsn := "postgres://postgres:postgres@localhost:5432/gophkeeper_test?sslmode=disable"

	dbConn, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	err = dbConn.Ping()
	require.NoError(t, err)

	// Очистка перед каждым запуском
	_, err = dbConn.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	require.NoError(t, err)

	return dbConn
}

// TestApplyMigrations_Success — тестирует успешное применение миграций
func TestApplyMigrations_Success(t *testing.T) {
	dbConn := setupTestDB(t)
	defer dbConn.Close()

	err := db.ApplyMigrations(dbConn)
	assert.NoError(t, err)

	var exists bool
	err = dbConn.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'users')").Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}
