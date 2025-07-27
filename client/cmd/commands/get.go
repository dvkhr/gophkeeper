package commands

import (
	"fmt"
	"strings"

	"github.com/dvkhr/gophkeeper/client/session"
	"github.com/urfave/cli/v2"
)

// NewGetCommand создаёт команду get
func NewGetCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Получить все сохранённые данные",
		Action: func(cCtx *cli.Context) error {
			client, err := session.LoadAuthenticatedClient(serverAddress)
			if err != nil {
				return err
			}
			defer client.Close()

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
				fmt.Printf("Данные:   %s\n", record.EncryptedData)
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
