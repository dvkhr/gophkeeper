package commands

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/internal/client"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/urfave/cli/v2"
)

// NewRegisterCommand создаёт команду register
func NewRegisterCommand(serverAddress string) *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Зарегистрировать нового пользователя",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "login", Aliases: []string{"l"}, Required: true},
			&cli.StringFlag{Name: "password", Aliases: []string{"p"}, Required: true},
		},
		Action: func(cCtx *cli.Context) error {
			login := cCtx.String("login")
			masterPassword := cCtx.String("password")

			logger.Logg.Debug("Начало регистрации", "login", login)

			salt, err := crypto.GenerateSalt()
			if err != nil {
				logger.Logg.Error("Не удалось сгенерировать соль", "error", err)
				return err
			}
			logger.Logg.Debug("Соль сгенерирована", "length", len(salt))

			key := crypto.DeriveKey(masterPassword, salt)
			logger.Logg.Debug("Ключ шифрования сгенерирован", "key_length", len(key))

			masterKeyHash := crypto.SHA256(key)

			client, err := client.NewClient(serverAddress, key)
			if err != nil {
				logger.Logg.Error("Не удалось создать gRPC-клиент", "error", err)
				return err
			}
			defer client.Close()

			logger.Logg.Info("Отправка запроса на регистрацию", "server", serverAddress)
			resp, err := client.Register(login, []byte(masterPassword))
			if err != nil {
				logger.Logg.Error("Регистрация не удалась", "login", login, "error", err)
				return err
			}
			logger.Logg.Debug("Регистрация успешна", "user_id", resp.UserId)

			session := &file.Data{
				Salt:          salt,
				MasterKeyHash: masterKeyHash,
				AccessToken:   resp.AccessToken,
				RefreshToken:  resp.RefreshToken,
			}

			if err := file.Save(session); err != nil {
				logger.Logg.Error("Не удалось сохранить сессию", "error", err)
				return fmt.Errorf("регистрация успешна, но не удалось сохранить сессию: %w", err)
			}
			logger.Logg.Debug("Полная сессия сохранена: salt, masterKeyHash, токены")

			fmt.Printf("Пользователь %s успешно зарегистрирован и авторизован\n", login)
			return nil
		},
	}
}
