// config/config.go

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// AuthConfig — конфигурация для JWT и refresh-токенов
type AuthConfig struct {
	JWTSecret           string `yaml:"jwt_secret"`
	JWTTTLHours         int    `yaml:"jwt_ttl_hours"`
	RefreshTokenTTLDays int    `yaml:"refresh_token_ttl_days"`
}

// Config — основная структура конфигурации приложения
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`

	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`

	Auth AuthConfig `yaml:"auth"`
}

// Load загружает конфигурацию из указанного YAML-файла
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
