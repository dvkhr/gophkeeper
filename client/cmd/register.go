package main

import (
	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/urfave/cli/v2"
)

func NewRegisterCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Зарегистрировать нового пользователя",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "login",
				Aliases:  []string{"l"},
				Usage:    "Логин пользователя",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"p"},
				Usage:    "Мастер-пароль",
				Required: true,
			},
		},
		Action: func(cCtx *cli.Context) error {
			login := cCtx.String("login")
			masterPassword := cCtx.String("password")

			// Генерируем соль
			salt, err := crypto.GenerateSalt()
			if err != nil {
				return err
			}

			// Сохраняем соль
			session := &file.Data{Salt: salt}
			if err := file.Save(session); err != nil {
				return err
			}

			// Генерируем ключ
			key := crypto.DeriveKey(masterPassword, salt)

			// Создаём клиент
			client, err := grpc.New("localhost:50051", key)
			if err != nil {
				return err
			}
			defer client.Close()

			// Регистрация
			resp, err := client.Register(login, []byte(masterPassword))
			if err != nil {
				return err
			}

			// Сохраняем токены
			return client.SetToken(resp.AccessToken, resp.RefreshToken)
		},
	}
}
