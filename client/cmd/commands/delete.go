package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/internal/client"
	"github.com/urfave/cli/v2"
)

// NewDeleteCommand создаёт команду delete.
func NewDeleteCommand(factory *client.Factory) *cli.Command {
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

			client, err := factory.NewAuthenticatedClient()
			if err != nil {
				return err
			}
			defer client.Close()

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
