package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config — это основная структура конфигурации
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`

	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`

	Auth struct {
		JWTSecret   string `yaml:"jwt_secret"`
		JWTTTLHours int    `yaml:"jwt_ttl_hours"`
	} `yaml:"auth"`
}

// Load загружает конфигурацию из указанного файла
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
