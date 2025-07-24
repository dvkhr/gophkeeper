package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "gophkeeper",
		Usage:   "Клиент для безопасного хранения данных",
		Version: fmt.Sprintf("%s (сборка: %s)", Version, BuildDate),
		Compiled: func() time.Time {
			if BuildDate == "unknown" {
				return time.Time{}
			}
			t, _ := time.Parse("2006-01-02T15:04:05Z", BuildDate)
			return t
		}(),
		Commands: []*cli.Command{
			NewVersionCommand(),
			NewRegisterCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
