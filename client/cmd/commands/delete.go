package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/internal/utils"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/urfave/cli/v2"
)

// NewDeleteCommand создаёт команду delete
func NewDeleteCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Удалить запись по ID",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Aliases:  []string{"i"},
				Required: true,
				Usage:    "ID записи, которую нужно удалить",
			},
		},
		Action: func(cCtx *cli.Context) error {
			id := cCtx.String("id")

			session, err := file.Load()
			if err != nil || session.AccessToken == "" {
				return fmt.Errorf("вы не авторизованы")
			}

			masterPassword, err := utils.ReadMasterPassword("Master-пароль: ")
			if err != nil {
				return fmt.Errorf("не удалось считать пароль: %w", err)
			}
			defer utils.ZeroBytes(masterPassword)

			key := crypto.DeriveKey(string(masterPassword), session.Salt)

			client, err := grpc.New(serverAddress, key)
			if err != nil {
				return err
			}
			defer client.Close()

			client.SetToken(session.AccessToken, session.RefreshToken)

			err = client.DoWithRetry(func() error {
				resp, err := client.DeleteData(id)
				if err != nil {
					return err
				}
				fmt.Printf("%s\n", resp.Message)
				return nil
			})

			return err
		},
	}
}
