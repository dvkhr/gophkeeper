// Package repository предоставляет реализацию репозиториев для работы с базой данных
// в GophKeeper. Поддерживает работу с пользователями, данными и refresh-токенами.
package repository

//godoc -http=:6060
//http://localhost:6060/pkg/github.com/dvkhr/gophkeeper/server/internal/repository/

// User представляет пользователя в системе
type User struct {
	ID           string
	Login        string
	PasswordHash string
	Status       string
	CreatedAt    int64
	UpdatedAt    int64
}

// DataRecord — модель данных пользователя, соответствует pb.DataRecord.
// Используется для хранения зашифрованных данных в базе данных.
type DataRecord struct {
	ID            string
	UserID        string
	Type          string
	EncryptedData []byte
	Metadata      map[string]string
	CreatedAt     int64
	UpdatedAt     int64
	Deleted       bool
}

// Repository — общий интерфейс для всех репозиториев
type Repository interface {
	UserRepository
	DataRepository
	TokenRepository
}
