package repository

import (
	"testing"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataRepository_SaveAndGet(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	dataRepo := NewDataRepository(db)

	login := "testuser"
	password := "hashedpass"

	// 1. Создаем пользователя
	userID, err := userRepo.CreateUser(login, password)
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	// 2. Подготавливаем тестовую запись
	record := &pb.DataRecord{
		Id:            "1",
		Type:          "loginpass",
		EncryptedData: []byte("encrypted_data"),
		Metadata:      map[string]string{"site": "example.com"},
	}

	// 3. Сохраняем данные
	err = dataRepo.SaveData(userID, record)
	require.NoError(t, err)

	// 4. Получаем данные
	records, err := dataRepo.GetAllData(userID)
	require.NoError(t, err)
	require.Len(t, records, 1)

	// 5. Проверяем данные
	assert.Equal(t, record.Id, records[0].Id)
	assert.Equal(t, record.Type, records[0].Type)
	assert.Equal(t, record.EncryptedData, records[0].EncryptedData)
	assert.Equal(t, record.Metadata, records[0].Metadata)
}

func TestDataRepository_GetAllData_Empty(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	dataRepo := NewDataRepository(db)

	login := "testuser"
	password := "hashedpass"

	// 1. Создаем пользователя
	userID, err := userRepo.CreateUser(login, password)
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	// 2. Запрашиваем данные — их ещё нет
	records, err := dataRepo.GetAllData(userID)
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestDataRepository_SaveData_Overwrite(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	dataRepo := NewDataRepository(db)

	login := "testuser"
	password := "hashedpass"

	userID, err := userRepo.CreateUser(login, password)
	require.NoError(t, err)

	record1 := &pb.DataRecord{
		Id:            "a1234567-bcde-8901-f234-567890ab",
		Type:          "loginpass",
		EncryptedData: []byte("encrypted_data_1"),
		Metadata:      map[string]string{"site": "example.com"},
	}

	record2 := &pb.DataRecord{
		Id:            record1.Id,
		Type:          "loginpass",
		EncryptedData: []byte("encrypted_data_2"),
		Metadata:      map[string]string{"site": "updated.com"},
	}

	// 1. Сохраняем первый раз
	err = dataRepo.SaveData(userID, record1)
	require.NoError(t, err)

	// 2. Обновляем запись
	err = dataRepo.SaveData(userID, record2)
	require.NoError(t, err)

	// 3. Читаем обратно
	records, err := dataRepo.GetAllData(userID)
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, record2.EncryptedData, records[0].EncryptedData)
	assert.Equal(t, record2.Metadata, records[0].Metadata)
}

func TestDataRepository_GetAllData_Multiple(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	dataRepo := NewDataRepository(db)

	login := "testuser"
	password := "hashedpass"

	userID, err := userRepo.CreateUser(login, password)
	require.NoError(t, err)

	rec1 := &pb.DataRecord{
		Id:            "1",
		Type:          "loginpass",
		EncryptedData: []byte("data1"),
		Metadata:      map[string]string{"site": "example.com"},
	}

	rec2 := &pb.DataRecord{
		Id:            "2",
		Type:          "card",
		EncryptedData: []byte("data2"),
		Metadata:      map[string]string{"bank": "bank.com"},
	}

	err = dataRepo.SaveData(userID, rec1)
	require.NoError(t, err)

	err = dataRepo.SaveData(userID, rec2)
	require.NoError(t, err)

	records, err := dataRepo.GetAllData(userID)
	require.NoError(t, err)
	assert.Len(t, records, 2)
}
