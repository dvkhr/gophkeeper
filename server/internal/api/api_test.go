package api

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var testDB *sql.DB

func setupTestServer(t *testing.T) *KeeperServer {
	logger.Logg = logger.NewTestLogger()

	dsn := "postgres://postgres:postgres@localhost:5432/gophkeeper_test?sslmode=disable"
	if testDB == nil {
		var err error
		testDB, err = sql.Open("pgx", dsn)
		require.NoError(t, err)

		//err = db.ApplyMigrations(testDB)
		//require.NoError(t, err)
		_, err = testDB.Exec(`
        DELETE FROM refresh_tokens;
        DELETE FROM user_data;
        DELETE FROM users;
    `)
		require.NoError(t, err)
	}

	_, err := testDB.Exec("DELETE FROM user_data; DELETE FROM users;")
	require.NoError(t, err)

	repo := repository.NewPostgresRepository(testDB)

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:           "test-secret",
			JWTTTLHours:         1,
			RefreshTokenTTLDays: 7,
		},
	}

	return NewKeeperServer(repo, cfg)
}

// успешная регистрация
func TestRegister_Success(t *testing.T) {
	server := setupTestServer(t)

	req := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("encrypted-pass-123"),
	}

	resp, err := server.Register(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

// повторная регистрация
func TestRegister_DuplicateLogin(t *testing.T) {
	server := setupTestServer(t)

	req := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("encrypted-pass-123"),
	}
	_, err := server.Register(context.Background(), req)
	require.NoError(t, err)

	_, err = server.Register(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

// успешный вход
func TestLogin_Success(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("encrypted-pass-123"),
	}
	_, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	loginReq := &pb.LoginRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("encrypted-pass-123"),
	}

	resp, err := server.Login(context.Background(), loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

// неверный пароль
func TestLogin_InvalidPassword(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("correct-pass"),
	}
	_, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	loginReq := &pb.LoginRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}

	_, err = server.Login(context.Background(), loginReq)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

// пользователь не существует.
func TestLogin_UserNotFound(t *testing.T) {
	server := setupTestServer(t)

	req := &pb.LoginRequest{
		Login:             "nonexistent",
		EncryptedPassword: []byte("any-pass"),
	}

	_, err := server.Login(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
