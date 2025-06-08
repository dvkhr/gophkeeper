package main

import (
	"fmt"
	"os"

	"github.com/dvkhr/gophkeeper/pkg/logger"
)

func main() {
	// Инициализируем логгер
	if err := logger.InitLogger("/home/max/go/src/GophKeeper/configs/logger.yml"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Тестовые сообщения
	logger.Logg.Info("Application started")
	logger.Logg.Debug("This is a debug message")
	logger.Logg.Warn("This is a warning message")
	logger.Logg.Error("This is an error message")

	// Пример маскировки чувствительных данных
	jsonData := `{"username": "admin", "password": "secret123"}`
	masked := logger.MaskSensitiveData(jsonData)
	logger.Logg.Info("Masked data", "data", masked)
}
