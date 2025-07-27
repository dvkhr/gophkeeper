package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/dvkhr/gophkeeper/client/session"
	"github.com/dvkhr/gophkeeper/pb"
	"github.com/urfave/cli/v2"
)

// NewGetCommand создаёт команду get
func NewGetCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Получить все сохранённые данные",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "id",
				Usage: "ID записи (опционально, для одной записи)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Путь для сохранения данных (для бинарных данных)",
			},
		},
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

			// для одной конкретной записи
			if cCtx.String("id") != "" {
				return printSingleRecord(cCtx, resp.Records)
			}

			// Вывод всех записей
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

				if cCtx.String("output") != "" || record.Type == "binary" {
					fmt.Printf("Данные:   (%d байт, тип %s) — используйте --output для сохранения\n", len(record.EncryptedData), record.Type)
				} else {
					fmt.Printf("Данные:   %s\n", record.EncryptedData)
				}

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

// printSingleRecord выводит одну запись, можно в файл для бинарных данных
func printSingleRecord(cCtx *cli.Context, records []*pb.DataRecord) error {
	var record *pb.DataRecord
	for _, r := range records {
		if r.Id == cCtx.String("id") {
			record = r
			break
		}
	}

	if record == nil {
		return fmt.Errorf("запись с ID %s не найдена", cCtx.String("id"))
	}

	outputPath := cCtx.String("output")
	if outputPath != "" {
		err := os.WriteFile(outputPath, record.EncryptedData, 0600)
		if err != nil {
			return fmt.Errorf("не удалось сохранить файл: %w", err)
		}
		fmt.Printf("Файл сохранён: %s (%d байт)\n", outputPath, len(record.EncryptedData))
		return nil
	}

	// Вывод в терминал (только для текстовых типов)
	if record.Type == "text" || record.Type == "loginpass" || record.Type == "card" {
		fmt.Printf("ID:       %s\n", record.Id)
		fmt.Printf("Тип:      %s\n", record.Type)
		fmt.Printf("Данные:   %s\n", record.EncryptedData)
		if len(record.Metadata) > 0 {
			fmt.Println("Метаданные:")
			for k, v := range record.Metadata {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
	} else {
		fmt.Printf("Тип: %s, размер: %d байт. Используйте --output для сохранения.\n", record.Type, len(record.EncryptedData))
	}

	return nil
}
