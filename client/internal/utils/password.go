package utils

import (
	"fmt"
	"syscall"

	"golang.org/x/term"
)

// ReadMasterPassword запрашивает мастер-пароль со скрытым вводом
func ReadMasterPassword(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return password, err
}

// ZeroBytes обнуляет байты в памяти для безопасности
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
