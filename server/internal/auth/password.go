package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword хэширует пароль с использованием bcrypt.
// Возвращает хэш и ошибку, если хэширование не удалось.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

// CheckPasswordHash проверяет, совпадает ли пароль с хэшем.
// Возвращает true, если пароль совпадает, и false в противном случае.
func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
