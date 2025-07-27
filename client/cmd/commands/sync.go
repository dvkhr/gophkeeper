package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/urfave/cli/v2"
)

// NewSyncCommand создаёт команду sync
func NewSyncCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "Синхронизировать данные c сервером",
		Action: func(cCtx *cli.Context) error {
			session, err := file.Load()
			if err != nil || session.AccessToken == "" {
				return fmt.Errorf("вы не авторизованы")
			}

			masterPassword := "master-pass" // заменить!
			key := crypto.DeriveKey(masterPassword, session.Salt)

			client, err := grpc.New(serverAddress, key)
			if err != nil {
				return err
			}
			defer client.Close()

			client.SetToken(session.AccessToken, session.RefreshToken)

			/// загрузка из локального хранилища?
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
				// сохранение в локальное хранилище?
				return nil
			})

			return err
		},
	}
}
