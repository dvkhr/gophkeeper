package crypto

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltSize   = 32    // длина соли
	Iterations = 10000 //количество итераций
	KeyLength  = 32    //длина ключа
)

// DeriveKey генерирует ключ из пароля и соли с помощью PBKDF2
func DeriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key(
		[]byte(password), // пароль
		salt,             // соль
		Iterations,       // количество итераций
		KeyLength,        // длина ключа
		sha256.New,       // хэш-функция
	)
}

// GenerateSalt генерирует случайную соль длиной SaltSize
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// SHA256 возвращает хеш SHA-256 от входных данных
func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
