package logger

import "log/slog"

// GetLogLevel преобразует строку уровня лога в slog.Level
func GetLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // По умолчанию используем Info
	}
}
