package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/urfave/cli/v2"
)

// NewLoginCommand создаёт команду login
func NewLoginCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Войти в аккаунт",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "login", Aliases: []string{"l"}, Required: true},
			&cli.StringFlag{Name: "password", Aliases: []string{"p"}, Required: true},
		},
		Action: func(cCtx *cli.Context) error {
			login := cCtx.String("login")
			password := cCtx.String("password")

			tempKey := crypto.DeriveKey(password, []byte("temp-salt"))
			client, err := grpc.New(serverAddress, tempKey)
			if err != nil {
				return err
			}
			defer client.Close()

			resp, err := client.Login(login, []byte(password))
			if err != nil {
				return err
			}

			session, err := file.Load()
			if err != nil {
				return fmt.Errorf("не удалось загрузить сессию: %w", err)
			}

			session.AccessToken = resp.AccessToken
			session.RefreshToken = resp.RefreshToken

			if err := file.Save(session); err != nil {
				return fmt.Errorf("не удалось сохранить сессию: %w", err)
			}

			logger.Logg.Info("Успешный вход", "login", login)
			fmt.Printf("Вы вошли как %s\n", login)
			return nil
		},
	}
}
