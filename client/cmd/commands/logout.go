package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/internal/client"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/urfave/cli/v2"
)

// NewLogoutCommand создаёт команду logout
func NewLogoutCommand(factory *client.Factory) *cli.Command {
	return &cli.Command{
		Name:  "logout",
		Usage: "Выйти из аккаунта",
		Action: func(cCtx *cli.Context) error {
			client, err := factory.NewAuthenticatedClient()
			if err != nil {
				return err
			}
			defer client.Close()

			session, err := file.Load()
			if err != nil || session.RefreshToken == "" {
				return fmt.Errorf("вы не авторизованы")
			}

			err = client.Logout(session.RefreshToken)
			if err != nil {
				logger.Logg.Warn("Не удалось отозвать refresh_token на сервере", "error", err)
			} else {
				logger.Logg.Info("refresh_token отозван на сервере")
			}

			fmt.Println("Вы вышли из аккаунта")
			return nil
		},
	}
}
