package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/dvkhr/gophkeeper/client/internal/client"
	"github.com/dvkhr/gophkeeper/pb"
	"github.com/urfave/cli/v2"
)

// NewAddCommand создаёт команду add
func NewAddCommand(factory *client.Factory) *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Добавить данные c метаинформацией",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "id", Required: true},
			&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Required: true},
			&cli.StringFlag{Name: "login"},
			&cli.StringFlag{Name: "password"},
			&cli.StringFlag{Name: "number"},
			&cli.StringFlag{Name: "expiry"},
			&cli.StringFlag{Name: "cvv"},
			&cli.StringFlag{Name: "content"},
			&cli.StringFlag{Name: "file", Usage: "Путь к файлу для загрузки"},
			&cli.StringSliceFlag{Name: "meta", Aliases: []string{"m"}},
		},
		Action: func(cCtx *cli.Context) error {
			client, err := factory.NewAuthenticatedClient()
			if err != nil {
				return err
			}
			defer client.Close()

			record, err := buildDataRecord(cCtx)
			if err != nil {
				return err
			}

			err = client.DoWithRetry(func() error {
				resp, err := client.StoreData(record)
				if err != nil {
					return err
				}
				fmt.Printf("Данные сохранены: %s\n", resp.Message)
				return nil
			})

			return err
		},
	}
}

// buildDataRecord — вспомогательная функция
func buildDataRecord(cCtx *cli.Context) (*pb.DataRecord, error) {
	if err := validateFlags(cCtx); err != nil {
		return nil, err
	}

	metadata := buildMetadata(cCtx)

	data, err := readData(cCtx)
	if err != nil {
		return nil, err
	}

	return &pb.DataRecord{
		Id:            cCtx.String("id"),
		Type:          cCtx.String("type"),
		EncryptedData: data,
		Metadata:      metadata,
	}, nil
}

// validateFlags проверяет, что для типа данных переданы нужные флаги
func validateFlags(cCtx *cli.Context) error {
	dataType := cCtx.String("type")
	id := cCtx.String("id")

	if id == "" {
		return fmt.Errorf("требуется --id")
	}

	switch dataType {
	case "loginpass":
		if cCtx.String("login") == "" || cCtx.String("password") == "" {
			return fmt.Errorf("для loginpass нужны --login и --password")
		}
	case "card":
		if cCtx.String("number") == "" || cCtx.String("expiry") == "" {
			return fmt.Errorf("для card нужны --number и --expiry")
		}
	case "text":
		if cCtx.String("content") == "" {
			return fmt.Errorf("для text нужен --content")
		}
	case "binary":
		if cCtx.String("file") == "" {
			return fmt.Errorf("для binary нужен --file")
		}
	default:
		return fmt.Errorf("неизвестный тип: %s", dataType)
	}

	return nil
}

// buildMetadata парсит флаг --meta в map[string]string
func buildMetadata(cCtx *cli.Context) map[string]string {
	metadata := make(map[string]string)
	for _, meta := range cCtx.StringSlice("meta") {
		for _, pair := range strings.Split(meta, ",") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				metadata[kv[0]] = kv[1]
			}
		}
	}
	return metadata
}

// readData читает данные в зависимости от типа
func readData(cCtx *cli.Context) ([]byte, error) {
	dataType := cCtx.String("type")

	if cCtx.String("file") != "" {
		data, err := os.ReadFile(cCtx.String("file"))
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
		}
		return data, nil
	}

	switch dataType {
	case "loginpass":
		return []byte(fmt.Sprintf("login:%s\npassword:%s",
			cCtx.String("login"), cCtx.String("password"))), nil
	case "card":
		return []byte(fmt.Sprintf("number:%s\nexpiry:%s\ncvv:%s",
			cCtx.String("number"), cCtx.String("expiry"), cCtx.String("cvv"))), nil
	case "text":
		return []byte(cCtx.String("content")), nil
	}

	return nil, fmt.Errorf("неожиданный тип: %s", dataType)
}
