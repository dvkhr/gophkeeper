package commands

import (
	"fmt"
	"strings"

	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/urfave/cli/v2"
)

// NewGetCommand создаёт команду get
func NewGetCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Получить все сохранённые данные",
		Action: func(cCtx *cli.Context) error {
			session, err := file.Load()
			if err != nil || session.AccessToken == "" {
				return fmt.Errorf("вы не авторизованы")
			}

			masterPassword := "master-pass" // заменить!!!
			key := crypto.DeriveKey(masterPassword, session.Salt)

			client, err := grpc.New(serverAddress, key)
			if err != nil {
				return err
			}
			defer client.Close()

			client.SetToken(session.AccessToken, session.RefreshToken)

			resp, err := client.GetData()
			if err != nil {
				return err
			}

			if len(resp.Records) == 0 {
				fmt.Println("Нет сохранённых данных")
				return nil
			}

			fmt.Printf("\n Найдено записей: %d\n", len(resp.Records))
			fmt.Println(strings.Repeat("─", 80))

			for i, record := range resp.Records {
				if i > 0 {
					fmt.Println(strings.Repeat("─", 80))
				}

				fmt.Printf("ID:       %s\n", record.Id)
				fmt.Printf("Тип:      %s\n", record.Type)

				data := string(record.EncryptedData)
				fmt.Printf("Данные:   %s\n", data)

				if len(record.Metadata) > 0 {
					fmt.Println("Метаданные:")
					for k, v := range record.Metadata {
						fmt.Printf("  %s: %s\n", k, v)
					}
				}
			}

			fmt.Println(strings.Repeat("─", 80))
			return nil
		},
	}
}
