package db_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Инициализируем логгер перед запуском тестов
	if err := logger.InitLogger("/home/max/go/src/GophKeeper/configs/logger.yaml"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestConnect_Success(t *testing.T) {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=gophkeeper_test sslmode=disable"
	dbConn, err := db.Connect(dsn)
	assert.NoError(t, err)
	assert.NotNil(t, dbConn)
	defer dbConn.Close()
}

func TestConnect_InvalidDSN(t *testing.T) {
	dsn := "invalid-dsn"
	dbConn, err := db.Connect(dsn)
	assert.Error(t, err)
	assert.Nil(t, dbConn)
}
