package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// NewVersionCommand возвращает команду version
func NewVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Показать версию и дату сборки",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Вывести в формате JSON",
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.Bool("json") {
				fmt.Printf(`{"version": "%s", "build_date": "%s"}`, Version, BuildDate)
			} else {
				fmt.Printf("GophKeeper CLI\n")
				fmt.Printf("Версия: %s\n", Version)
				fmt.Printf("Дата сборки: %s\n", BuildDate)
			}
			return nil
		},
	}
}
