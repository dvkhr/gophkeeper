package logger

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config — конфигурация логгера
type Config struct {
	LogLevel      string `yaml:"log_level"`
	ConsoleFormat string `yaml:"console_format"`
	FileFormat    string `yaml:"file_format"`
	Destination   string `yaml:"destination"`
	FilePattern   string `yaml:"file_pattern"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GenerateFileName генерирует имя файла на основе шаблона даты
func GenerateFileName(pattern string) string {
	now := time.Now()
	fileName := now.Format(pattern)
	dir := "./logs"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}
	return fileName
}
