package config

import (
	"flag"
	"os"
)

var (
	ConfigFile string
)

func ParseFlags() {
	flag.StringVar(&ConfigFile, "config", getDefaultConfigPath(), "Path to config file")

	flag.Parse()
}

func getDefaultConfigPath() string {
	if env := os.Getenv("CONFIG_FILE"); env != "" {
		return env
	}
	return "/home/max/go/src/GophKeeper/configs/config.yaml"
}
