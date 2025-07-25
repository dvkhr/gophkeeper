package session

import (
	"fmt"

	"github.com/dvkhr/gophkeeper/client/grpc"
	"github.com/dvkhr/gophkeeper/client/storage/file"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
)

// Manager управляет сессией клиента: загрузка соли, ввод пароля, создание gRPC-клиента
type Manager struct {
	address string
}

// NewManager создаёт новый менеджер сессии
func NewManager(address string) *Manager {
	return &Manager{address: address}
}

// NewClient создаёт и настраивает gRPC-клиент
func (m *Manager) NewClient() (*grpc.Client, error) {
	session, err := file.Load()
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить сессию: %w", err)
	}

	if session.Salt == nil {
		return nil, fmt.Errorf("соль не найдена. Выполните регистрацию")
	}

	masterPassword := "master-pass-placeholder" // ← заменим на realPassword()

	key := crypto.DeriveKey(masterPassword, session.Salt)

	client, err := grpc.New(m.address, key)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать gRPC-клиент: %w", err)
	}

	if session.AccessToken != "" && session.RefreshToken != "" {
		_ = client.SetToken(session.AccessToken, session.RefreshToken)
	}

	return client, nil
}

// NewClientWithPassword создаёт клиент, используя переданный мастер-пароль
func (m *Manager) NewClientWithPassword(masterPassword string) (*grpc.Client, error) {
	session, err := file.Load()
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить сессию: %w", err)
	}

	if session.Salt == nil {
		return nil, fmt.Errorf("соль не найдена")
	}

	key := crypto.DeriveKey(masterPassword, session.Salt)
	client, err := grpc.New(m.address, key)
	if err != nil {
		return nil, err
	}

	// Если токены есть — восстанавливаем
	if session.AccessToken != "" && session.RefreshToken != "" {
		_ = client.SetToken(session.AccessToken, session.RefreshToken)
	}

	return client, nil
}
