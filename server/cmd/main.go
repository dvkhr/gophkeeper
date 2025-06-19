package main

import (
	"fmt"
	"os"

	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/db"
)

const dsn = "host=localhost port=5432 user=postgres password=postgres dbname=gophkeeper sslmode=disable"

func main() {
	// Инициализируем логгер
	if err := logger.InitLogger("/home/max/go/src/GophKeeper/configs/logger.yml"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	dbConn, err := db.Connect(dsn)
	if err != nil {
		logger.Logg.Error("Failed to connect to DB", "error", err)
		return
	}
	defer dbConn.Close()

	if err := db.ApplyMigrations(dbConn); err != nil {
		logger.Logg.Error("Failed to apply migrations", "error", err)
		return
	}

	logger.Logg.Info("Database is ready. Starting server...")

}
