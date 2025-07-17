package repository

// User представляет пользователя в системе
type User struct {
	ID           string
	Login        string
	PasswordHash string
	Status       string
	CreatedAt    int64
	UpdatedAt    int64
}

// DataRecord — модель данных пользователя, соответствует pb.DataRecord
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
