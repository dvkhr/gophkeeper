package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/dvkhr/gophkeeper/pkg/logger"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func ApplyMigrations(db *sql.DB) error {
	logger.Logg.Info("Running database migrations...")

	files, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		logger.Logg.Error("Failed to read migration files", "error", err)
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}

		migrationName := file.Name()
		logger.Logg.Info("Applying migration", "name", migrationName)

		content, err := fs.ReadFile(migrationFiles, "migrations/"+migrationName)
		if err != nil {
			logger.Logg.Error("Failed to read migration file", "name", migrationName, "error", err)
			return fmt.Errorf("failed to read %s: %w", migrationName, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			logger.Logg.Error("Failed to apply migration", "name", migrationName, "error", err)
			return fmt.Errorf("failed to apply %s: %w", migrationName, err)
		}
	}

	logger.Logg.Info("All migrations applied successfully.")
	return nil
}
