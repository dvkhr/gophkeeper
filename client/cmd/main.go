package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dvkhr/gophkeeper/client/cmd/commands"
	"github.com/dvkhr/gophkeeper/client/internal/config"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/urfave/cli/v2"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func main() {
	if err := logger.InitLogger("/home/max/go/src/GophKeeper/configs/logger.yaml"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	var flagServer string

	app := &cli.App{
		Name:    "gophkeeper-client",
		Usage:   "CLI-клиент для GophKeeper",
		Version: fmt.Sprintf("%s (сборка: %s)", Version, BuildDate),
		Compiled: func() time.Time {
			if BuildDate == "unknown" {
				return time.Time{}
			}
			t, _ := time.Parse("2006-01-02T15:04:05Z", BuildDate)
			return t
		}(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "server",
				Aliases:     []string{"s"},
				Usage:       "Адрес gRPC-сервера",
				Destination: &flagServer,
				EnvVars:     []string{"GK_SERVER"},
			},
		},
		Before: func(cCtx *cli.Context) error {
			// Загружаем конфиг из: флаг > env > файл
			cfg := config.Load(flagServer)

			for i, cmd := range cCtx.App.Commands {
				switch cmd.Name {
				case "register":
					cCtx.App.Commands[i] = commands.NewRegisterCommand(cfg.Server.Address)
				case "add":
					cCtx.App.Commands[i] = commands.NewAddCommand(cfg.Server.Address)
				case "login":
					cCtx.App.Commands[i] = commands.NewLoginCommand(cfg.Server.Address)
				case "logout":
					cmd.Action = commands.NewLogoutCommand(cfg.Server.Address).Action
				case "get":
					cCtx.App.Commands[i] = commands.NewGetCommand(cfg.Server.Address)
				case "delete":
					cCtx.App.Commands[i] = commands.NewDeleteCommand(cfg.Server.Address)
				case "sync":
					cCtx.App.Commands[i] = commands.NewSyncCommand(cfg.Server.Address)
				}
			}
			return nil
		},
		Commands: []*cli.Command{
			NewVersionCommand(),
			{Name: "register"},
			{Name: "add"},
			{Name: "login"},
			{Name: "logout"},
			{Name: "get"},
			{Name: "delete"},
			{Name: "sync"},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
