package crypto

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltSize = 32
	Iter     = 10000
	KeyLen   = 32 // AES-256
)

// DeriveKey генерирует ключ из пароля и соли
func DeriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, Iter, KeyLen, sha256.New)
}

// GenerateSalt генерирует случайную соль
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
