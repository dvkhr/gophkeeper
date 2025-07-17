// pkg/repository/user_test.go

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndGet(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	login := "testuser"
	password := "hashedpass"

	userID, err := repo.CreateUser(login, password)
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	user, err := repo.GetUserByLogin(login)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, login, user.Login)
	assert.Equal(t, password, user.PasswordHash)
}

func TestUserRepository_GetUser_NotFound(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	user, err := repo.GetUserByLogin("notexists")
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_CreateUser_DuplicateLogin(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	login := "duplicateuser"
	password := "pass123"

	// Первый раз создаём пользователя
	_, err := repo.CreateUser(login, password)
	require.NoError(t, err)

	// Второй раз — ожидаем ошибку
	_, err = repo.CreateUser(login, password)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
}

func TestUserRepository_GetUser_NotActive(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	login := "inactiveuser"
	password := "pass123"

	userID, err := repo.CreateUser(login, password)
	require.NoError(t, err)

	// Обновляем статус на 'blocked'
	_, err = db.ExecContext(context.Background(),
		`UPDATE users SET status = 'blocked' WHERE id = $1`, userID)
	require.NoError(t, err)

	user, err := repo.GetUserByLogin(login)
	assert.NoError(t, err)
	assert.Nil(t, user)
}
