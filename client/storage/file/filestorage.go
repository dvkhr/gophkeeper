package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Data — данные, которые хранятся в файле
type Data struct {
	Salt          []byte `json:"salt,omitempty"`
	AccessToken   string `json:"access_token,omitempty"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	MasterKeyHash []byte `json:"master_key_hash,omitempty"`
}

// getPath возвращает путь к файлу данных
func getPath() string {
	var dir string
	if runtime.GOOS == "windows" {
		dir = os.Getenv("APPDATA")
	} else {
		home := os.Getenv("HOME")
		dir = filepath.Join(home, ".config")
		// Создаём директорию, если её нет
		_ = os.MkdirAll(dir, 0700)
	}

	return filepath.Join(dir, ".gophkeeper.json")
}

// Save сохраняет данные в файл с правами 0600
func Save(data *Data) error {
	path := getPath()
	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, fileData, 0600)
}

// Load загружает данные из файла
func Load() (*Data, error) {
	path := getPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Файл ещё не создан
		return &Data{}, nil
	}

	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data Data
	err = json.Unmarshal(fileData, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// Clear удаляет файл
func Clear() error {
	path := getPath()
	return os.Remove(path)
}

// Delete удаляет файл сессии
func Delete() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(configDir, ".gophkeeper.json")

	return os.Remove(path)
}
