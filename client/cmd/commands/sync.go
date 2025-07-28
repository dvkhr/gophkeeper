package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/internal/client"
	"github.com/dvkhr/gophkeeper/pb"
	"github.com/urfave/cli/v2"
)

// NewSyncCommand создаёт команду sync
func NewSyncCommand(factory *client.Factory) *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "Синхронизировать данные с сервером",
		Action: func(cCtx *cli.Context) error {
			client, err := factory.NewAuthenticatedClient()
			if err != nil {
				return err
			}
			defer client.Close()

			// Локальное хранилище???
			var records []*pb.DataRecord

			err = client.DoWithRetry(func() error {
				resp, err := client.SyncData(records)
				if err != nil {
					return err
				}

				if len(resp.Records) == 0 {
					fmt.Println("Сервер не вернул новых данных")
					return nil
				}

				fmt.Printf("Получено %d записей с сервера:\n", len(resp.Records))
				for _, record := range resp.Records {
					fmt.Printf("  - ID: %s, Тип: %s\n", record.Id, record.Type)
				}

				return nil
			})

			return err
		},
	}
}
