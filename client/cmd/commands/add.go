package commands

import (
	"fmt"
	"strings"

	"github.com/dvkhr/gophkeeper/client/session"
	"github.com/dvkhr/gophkeeper/pb"
	"github.com/urfave/cli/v2"
)

// NewAddCommand создаёт команду add
func NewAddCommand(serverAddress string) *cli.Command {
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
			&cli.StringSliceFlag{Name: "meta", Aliases: []string{"m"}},
		},
		Action: func(cCtx *cli.Context) error {
			client, err := session.LoadAuthenticatedClient(serverAddress)
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
	dataType := cCtx.String("type")
	id := cCtx.String("id")

	switch dataType {
	case "loginpass":
		if cCtx.String("login") == "" || cCtx.String("password") == "" {
			return nil, fmt.Errorf("для loginpass нужны --login и --password")
		}
	case "card":
		if cCtx.String("number") == "" || cCtx.String("expiry") == "" {
			return nil, fmt.Errorf("для card нужны --number и --expiry")
		}
	case "text":
		if cCtx.String("content") == "" {
			return nil, fmt.Errorf("для text нужен --content")
		}
	default:
		return nil, fmt.Errorf("неизвестный тип: %s", dataType)
	}

	metadata := make(map[string]string)
	for _, meta := range cCtx.StringSlice("meta") {
		for _, pair := range strings.Split(meta, ",") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				metadata[kv[0]] = kv[1]
			}
		}
	}

	var data string
	switch dataType {
	case "loginpass":
		data = fmt.Sprintf("login:%s\npassword:%s", cCtx.String("login"), cCtx.String("password"))
	case "card":
		data = fmt.Sprintf("number:%s\nexpiry:%s\ncvv:%s", cCtx.String("number"), cCtx.String("expiry"), cCtx.String("cvv"))
	case "text":
		data = cCtx.String("content")
	}

	return &pb.DataRecord{
		Id:            id,
		Type:          dataType,
		EncryptedData: []byte(data),
		Metadata:      metadata,
	}, nil
}
