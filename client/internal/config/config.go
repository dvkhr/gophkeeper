package config

import (
	"os"

	"github.com/dvkhr/gophkeeper/pkg/logger"
	"gopkg.in/yaml.v3"
)

const configFilePath = "/home/max/go/src/GophKeeper/configs/client.yaml"

type Config struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
}

func Load(flagAddress string) *Config {
	cfg := &Config{}

	loadFromFile(cfg)

	if envAddr := os.Getenv("GK_SERVER"); envAddr != "" {
		if err := ValidateServerAddress(envAddr); err != nil {
			logger.Logg.Warn("Invalid server address in GK_SERVER, skipped",
				"error", err)
		} else {
			cfg.Server.Address = envAddr
			logger.Logg.Info("Конфиг: адрес из переменной окружения", "GK_SERVER", envAddr)
		}
	}

	if flagAddress != "" {
		if err := ValidateServerAddress(flagAddress); err != nil {
			logger.Logg.Error("Invalid server address in --server flag",
				"error", err,
				"address", flagAddress)
		} else {
			cfg.Server.Address = flagAddress
			logger.Logg.Info("Конфиг: адрес из флага --server", "address", flagAddress)
		}
	}

	if cfg.Server.Address == "" {
		cfg.Server.Address = "localhost:50051"
		logger.Logg.Info("Конфиг: использован адрес по умолчанию", "address", cfg.Server.Address)
	} else {
		if err := ValidateServerAddress(cfg.Server.Address); err != nil {
			logger.Logg.Warn("Invalid server address from config file, using default",
				"error", err)
			cfg.Server.Address = "localhost:50051"
		}
	}

	return cfg
}

// loadFromFile загружает конфигурацию из YAML-файла, если он существует.
// Игнорирует все ошибки, логгируя их на уровне WARN.
// Не возвращает ошибку, так как файл не обязателен.
func loadFromFile(cfg *Config) {

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		logger.Logg.Warn("Конфигурационный файл не найден", "path", configFilePath)
		return
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		logger.Logg.Warn("Не удалось прочитать конфигурационный файл",
			"path", configFilePath,
			"error", err)
		return
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		logger.Logg.Warn("Не удалось разобрать YAML-конфиг",
			"path", configFilePath,
			"error", err)
		return
	}

	logger.Logg.Info("Конфигурация успешно загружена из файла", "path", configFilePath)
}
