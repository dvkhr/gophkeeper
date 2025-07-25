package main

import (
	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/urfave/cli/v2"
)

// NewRegisterCommand создаёт команду register
func NewRegisterCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Зарегистрировать нового пользователя",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "login", Aliases: []string{"l"}, Required: true},
			&cli.StringFlag{Name: "password", Aliases: []string{"p"}, Required: true},
		},
		Action: func(cCtx *cli.Context) error {
			login := cCtx.String("login")
			masterPassword := cCtx.String("password")

			salt, err := crypto.GenerateSalt()
			if err != nil {
				return err
			}

			if err := file.Save(&file.Data{Salt: salt}); err != nil {
				return err
			}

			key := crypto.DeriveKey(masterPassword, salt)

			client, err := grpc.New(serverAddress, key)
			if err != nil {
				return err
			}
			defer client.Close()

			resp, err := client.Register(login, []byte(masterPassword))
			if err != nil {
				return err
			}

			return client.SetToken(resp.AccessToken, resp.RefreshToken)
		},
	}
}
