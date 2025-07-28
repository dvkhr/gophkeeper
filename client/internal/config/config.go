package config

import (
	"os"

	"github.com/dvkhr/gophkeeper/pkg/logger"
)

type Config struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
}

func Load(flagAddress string) *Config {
	cfg := &Config{}

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
