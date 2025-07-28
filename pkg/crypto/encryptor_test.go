package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// проверка шифрования
func TestEncryptor_EncryptDecrypt_Success(t *testing.T) {
	key := []byte("this-is-32-byte-key-for-aes-256!")
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	plaintext := []byte("secret data for gophkeeper")

	ciphertext, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	assert.Greater(t, len(ciphertext), len(plaintext)) // nonce + overhead

	decrypted, err := encryptor.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

// короткий ключ
func TestEncryptor_InvalidKeySize(t *testing.T) {
	_, err := NewEncryptor([]byte("short"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid key size")
}

// чужой ключ
func TestEncryptor_DecryptWithWrongKey(t *testing.T) {
	key1 := []byte("this-is-32-byte-key-for-aes-256!")
	key2 := []byte("another-32-byte-key-for-testing!")

	encryptor1, _ := NewEncryptor(key1)
	encryptor2, _ := NewEncryptor(key2)

	plaintext := []byte("secret data")

	ciphertext, err := encryptor1.Encrypt(plaintext)
	require.NoError(t, err)

	_, err = encryptor2.Decrypt(ciphertext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decryption failed")
}
