package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// сохранение и загрузка всех полей
func TestSaveAndLoad(t *testing.T) {
	testData := &Data{
		Salt:         []byte("test-salt-1234567890123456789012"),
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
	}

	filePath := getPath()

	_ = os.Remove(filePath)

	err := Save(testData)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(filePath)
	}()

	_, err = os.Stat(filePath)
	require.NoError(t, err)

	info, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, testData.Salt, loaded.Salt)
	assert.Equal(t, testData.AccessToken, loaded.AccessToken)
	assert.Equal(t, testData.RefreshToken, loaded.RefreshToken)
}

// пустая структура при отсутствии файла
func TestLoad_NotExists(t *testing.T) {
	loaded, err := Load()
	require.NoError(t, err)
	assert.Nil(t, loaded.Salt)
	assert.Empty(t, loaded.AccessToken)
	assert.Empty(t, loaded.RefreshToken)
}

// удаление файла
func TestClear(t *testing.T) {
	testData := &Data{
		Salt:         []byte("salt"),
		AccessToken:  "access",
		RefreshToken: "refresh",
	}
	err := Save(testData)
	require.NoError(t, err)

	_, err = os.Stat(getPath())
	require.NoError(t, err)

	err = Clear()
	require.NoError(t, err)

	_, err = os.Stat(getPath())
	assert.True(t, os.IsNotExist(err))
}

// пустые данные
func TestSave_EmptyData(t *testing.T) {
	err := Save(&Data{})
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)
	assert.Nil(t, loaded.Salt)
	assert.Empty(t, loaded.AccessToken)
	assert.Empty(t, loaded.RefreshToken)
}
