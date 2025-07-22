// pkg/logger/init.go

package logger

import (
	"fmt"
	"io"
	"os"

	"log/slog"

	"github.com/dvkhr/gophkeeper/pkg/logger/handler"
)

// InitLogger инициализирует логгер из конфигурации.
func InitLogger(configPath string) error {
	var err error
	loggerOnce.Do(func() {
		// Загружаем конфигурацию
		cfg, loadErr := LoadConfig(configPath)
		if loadErr != nil {
			err = fmt.Errorf("failed to load logger config: %w", loadErr)
			return
		}

		// Получаем параметры логгера из конфигурации
		logLevel := cfg.LogLevel
		consoleFormat := cfg.ConsoleFormat
		fileFormat := cfg.FileFormat
		destination := cfg.Destination
		filePattern := cfg.FilePattern

		level := GetLogLevel(logLevel)
		if level == slog.Level(-999) {
			Logg = &Logger{
				logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
			}
			return
		}

		var handlers []slog.Handler

		// Обработчик для терминала
		if destination == "console" || destination == "both" {
			var consoleHandler slog.Handler
			switch consoleFormat {
			case "text":
				consoleHandler = handler.NewTextHandler(os.Stdout, level)
			case "json":
				consoleHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
			default:
				consoleHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
			}
			handlers = append(handlers, consoleHandler)
		}

		// Обработчик для файла
		if destination == "file" || destination == "both" {
			fileName := GenerateFileName(filePattern)
			file, openErr := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if openErr != nil {
				fmt.Printf("Failed to open log file: %v\n", openErr)
			} else {
				var fileHandler slog.Handler
				switch fileFormat {
				case "text":
					fileHandler = handler.NewTextHandler(file, level)
				case "json":
					fileHandler = slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level})
				default:
					fileHandler = slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level})
				}
				handlers = append(handlers, fileHandler)
			}
		}

		// Если нет обработчиков, используем консоль по умолчанию
		if len(handlers) == 0 {
			handlers = append(handlers, slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
		}

		// Инициализируем логгер
		Logg = &Logger{
			logger: slog.New(handler.NewMultiHandler(handlers...)),
		}
	})
	return err
}

// InitTestLogger инициализирует тестовый логгер с уровнем логгирования "none".
func InitTestLogger() error {
	var err error
	loggerOnce.Do(func() {
		// Уровень логгирования "none" — отключаем все логи
		level := slog.Level(-999)

		// Создаем обработчик, который игнорирует все логи
		handler := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: level})

		// Инициализируем глобальный логгер
		Logg = &Logger{
			logger: slog.New(handler),
		}
	})
	return err
}

// NewTestLogger создаёт новый логгер для тестов
func NewTestLogger() *Logger {
	return &Logger{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}
