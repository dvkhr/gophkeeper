// Package session отвечает за управление сессией клиента.
// Данные сохраняются в файле `~/.config/.gophkeeper.json`.

package session

import "github.com/dvkhr/gophkeeper/client/storage/file"

// Data — данные сессии
type Data struct {
	Salt          []byte
	AccessToken   string
	RefreshToken  string
	MasterKeyHash []byte
}

// Manager управляет сессией клиента: загрузка соли, ввод пароля, создание gRPC-клиента
type Manager struct{}

// NewManager создаёт новый менеджер сессии
func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Load() (*Data, error) {
	data, err := file.Load()
	if err != nil {
		return nil, err
	}
	return &Data{
		Salt:          data.Salt,
		AccessToken:   data.AccessToken,
		RefreshToken:  data.RefreshToken,
		MasterKeyHash: data.MasterKeyHash,
	}, nil
}

func (m *Manager) Save(data *Data) error {
	return file.Save(&file.Data{
		Salt:          data.Salt,
		AccessToken:   data.AccessToken,
		RefreshToken:  data.RefreshToken,
		MasterKeyHash: data.MasterKeyHash,
	})
}

// проверяет наличие (не валидность) access-токена
func (m *Manager) IsAuthenticated() (bool, error) {
	s, err := m.Load()
	if err != nil {
		return false, err
	}
	return s.AccessToken != "", nil
}
