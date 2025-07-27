package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/urfave/cli/v2"
)

// NewLogoutCommand создаёт команду logout
func NewLogoutCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "logout",
		Usage: "Выйти из аккаунта",
		Action: func(cCtx *cli.Context) error {
			session, err := file.Load()
			if err != nil || session.RefreshToken == "" {
				return fmt.Errorf("вы не авторизованы")
			}

			masterPassword := "master-pass" //заменить!!!
			key := crypto.DeriveKey(masterPassword, session.Salt)

			client, err := grpc.New(serverAddress, key)
			if err != nil {
				return err
			}
			defer client.Close()

			client.SetToken(session.AccessToken, session.RefreshToken)

			err = client.Logout(session.RefreshToken)
			if err != nil {
				logger.Logg.Warn("Не удалось отозвать refresh_token на сервере", "error", err)
			} else {
				logger.Logg.Info("refresh_token отозван на сервере")
			}

			err = file.Delete()
			if err != nil {
				return fmt.Errorf("не удалось удалить сессию: %w", err)
			}

			fmt.Println("Вы вышли из аккаунта")
			return nil
		},
	}
}
