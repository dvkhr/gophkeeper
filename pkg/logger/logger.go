package logger

import (
	"log/slog"
	"sync"
)

type Logger struct {
	logger *slog.Logger
}

var (
	Logg       *Logger
	loggerOnce sync.Once
)

func (l *Logger) Info(msg string, attrs ...any) {
	l.logger.Info(msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...any) {
	l.logger.Warn(msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...any) {
	l.logger.Error(msg, attrs...)
}

func (l *Logger) Debug(msg string, attrs ...any) {
	l.logger.Debug(msg, attrs...)
}
