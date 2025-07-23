package api

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/auth"
	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/db"
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

		err = db.ApplyMigrations(testDB)
		require.NoError(t, err)
	}

	_, err := testDB.Exec(`
        DELETE FROM refresh_tokens;
        DELETE FROM user_data;
        DELETE FROM users;
    `)
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

// пользователь не существует
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

// успешное сохранение
func TestStoreData_Success(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	record := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("encrypted-data"),
		Metadata:      map[string]string{"site": "example.com"},
	}
	req := &pb.StoreDataRequest{Record: record}

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	resp, err := server.StoreData(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Data stored successfully", resp.Message)
}

// изменение
func TestStoreData_Update(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	record := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("data-v1"),
	}
	req := &pb.StoreDataRequest{Record: record}

	ctx := auth.WithUserID(context.Background(), userID.UserID)
	_, err = server.StoreData(ctx, req)
	require.NoError(t, err)

	record.EncryptedData = []byte("data-v2")
	_, err = server.StoreData(ctx, req)
	require.NoError(t, err)

	getResp, err := server.GetData(ctx, &pb.GetDataRequest{})
	require.NoError(t, err)
	require.Len(t, getResp.Records, 1)
	assert.Equal(t, []byte("data-v2"), getResp.Records[0].EncryptedData)
}

// пустой id
func TestStoreData_EmptyId(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	record := &pb.DataRecord{
		Id:            "",
		Type:          "loginpass",
		EncryptedData: []byte("data"),
	}
	req := &pb.StoreDataRequest{Record: record}

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	resp, err := server.StoreData(ctx, req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "Record ID is required")
	assert.Nil(t, resp)
}

// пустая запись
func TestStoreData_RecordNil(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	req := &pb.StoreDataRequest{Record: nil}
	ctx := auth.WithUserID(context.Background(), userID.UserID)

	resp, err := server.StoreData(ctx, req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "Record is required")
	assert.Nil(t, resp)
}

// неавторизированный пользователь
func TestStoreData_Unauthenticated(t *testing.T) {
	server := setupTestServer(t)

	req := &pb.StoreDataRequest{
		Record: &pb.DataRecord{
			Id:   "record-1",
			Type: "loginpass",
		},
	}

	ctx := context.Background()

	resp, err := server.StoreData(ctx, req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Nil(t, resp)
}

// нет данных
func TestGetData_Empty(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)
	resp, err := server.GetData(ctx, &pb.GetDataRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Records)
}

// есть данные
func TestGetData_WithRecords(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	record := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("encrypted-data"),
		Metadata:      map[string]string{"site": "example.com"},
	}
	storeReq := &pb.StoreDataRequest{Record: record}

	ctx := auth.WithUserID(context.Background(), userID.UserID)
	_, err = server.StoreData(ctx, storeReq)
	require.NoError(t, err)

	resp, err := server.GetData(ctx, &pb.GetDataRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Records, 1)
	assert.Equal(t, record.Id, resp.Records[0].Id)
	assert.Equal(t, record.Type, resp.Records[0].Type)
	assert.Equal(t, record.EncryptedData, resp.Records[0].EncryptedData)
	assert.Equal(t, record.Metadata, resp.Records[0].Metadata)
}

// неавторизированный пользователь
func TestGetData_Unauthenticated(t *testing.T) {
	server := setupTestServer(t)

	ctx := context.Background()

	resp, err := server.GetData(ctx, &pb.GetDataRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Nil(t, resp)
}

// без удаленных записей
func TestGetData_DoesNotReturnDeleted(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	record := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("data"),
	}
	_, err = server.StoreData(ctx, &pb.StoreDataRequest{Record: record})
	require.NoError(t, err)

	_, err = server.DeleteData(ctx, &pb.DeleteDataRequest{Id: "record-1"})
	require.NoError(t, err)

	resp, err := server.GetData(ctx, &pb.GetDataRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Records) // должна быть пустая
}

// пустой запрос
func TestSyncData_Empty(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)
	resp, err := server.SyncData(ctx, &pb.SyncRequest{Records: []*pb.DataRecord{}})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Records)
}

// синхронизация с данными
func TestSyncData_NewRecords(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	records := []*pb.DataRecord{
		{
			Id:            "record-1",
			Type:          "loginpass",
			EncryptedData: []byte("data-1"),
			Metadata:      map[string]string{"site": "site1.com"},
		},
		{
			Id:            "record-2",
			Type:          "card",
			EncryptedData: []byte("data-2"),
			Metadata:      map[string]string{"site": "site2.com"},
		},
	}

	ctx := auth.WithUserID(context.Background(), userID.UserID)
	resp, err := server.SyncData(ctx, &pb.SyncRequest{Records: records})
	require.NoError(t, err)
	require.Len(t, resp.Records, 2)

	ids := make(map[string]bool)
	for _, r := range resp.Records {
		ids[r.Id] = true
	}
	assert.True(t, ids["record-1"])
	assert.True(t, ids["record-2"])
}

// изменение записи
func TestSyncData_UpdateRecords(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	oldRecord := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("old-data"),
	}
	_, err = server.SyncData(ctx, &pb.SyncRequest{Records: []*pb.DataRecord{oldRecord}})
	require.NoError(t, err)

	updatedRecord := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("new-data"),
	}
	resp, err := server.SyncData(ctx, &pb.SyncRequest{Records: []*pb.DataRecord{updatedRecord}})
	require.NoError(t, err)
	require.Len(t, resp.Records, 1)
	assert.Equal(t, []byte("new-data"), resp.Records[0].EncryptedData)
}

// невалидные записи
func TestSyncData_SkipInvalidRecords(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	records := []*pb.DataRecord{
		nil,
		{Id: "", Type: "loginpass"},
		{Id: "record-1", Type: "card", EncryptedData: []byte("valid-data")},
	}

	resp, err := server.SyncData(ctx, &pb.SyncRequest{Records: records})
	require.NoError(t, err)
	require.Len(t, resp.Records, 1)
	assert.Equal(t, "record-1", resp.Records[0].Id)
}

// успешное удаление
func TestDeleteData_Success(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	record := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("data"),
	}
	_, err = server.StoreData(ctx, &pb.StoreDataRequest{Record: record})
	require.NoError(t, err)

	resp, err := server.DeleteData(ctx, &pb.DeleteDataRequest{Id: "record-1"})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Data deleted successfully", resp.Message)

	getResp, err := server.GetData(ctx, &pb.GetDataRequest{})
	require.NoError(t, err)
	require.Empty(t, getResp.Records)
}

// пустой id
func TestDeleteData_EmptyId(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	resp, err := server.DeleteData(ctx, &pb.DeleteDataRequest{Id: ""})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "Record ID is required")
	assert.Nil(t, resp)
}

// неавторизированный пользователь
func TestDeleteData_Unauthenticated(t *testing.T) {
	server := setupTestServer(t)

	ctx := context.Background()

	resp, err := server.DeleteData(ctx, &pb.DeleteDataRequest{Id: "record-1"})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Nil(t, resp)
}

// запись не существует
func TestDeleteData_NotFound(t *testing.T) {
	server := setupTestServer(t)

	registerReq := &pb.RegisterRequest{
		Login:             "testuser",
		EncryptedPassword: []byte("pass"),
	}
	registerResp, err := server.Register(context.Background(), registerReq)
	require.NoError(t, err)

	userID, err := auth.ParseToken(*server.cfg, registerResp.AccessToken)
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID.UserID)

	resp, err := server.DeleteData(ctx, &pb.DeleteDataRequest{Id: "nonexistent"})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "Data not found or access denied")
	assert.Nil(t, resp)
}

// удаление чужой записи
func TestDeleteData_OtherUserRecord(t *testing.T) {
	server := setupTestServer(t)

	registerReq1 := &pb.RegisterRequest{
		Login:             "user1",
		EncryptedPassword: []byte("pass"),
	}
	registerResp1, err := server.Register(context.Background(), registerReq1)
	require.NoError(t, err)

	userID1, err := auth.ParseToken(*server.cfg, registerResp1.AccessToken)
	require.NoError(t, err)

	ctx1 := auth.WithUserID(context.Background(), userID1.UserID)

	record := &pb.DataRecord{
		Id:            "record-1",
		Type:          "loginpass",
		EncryptedData: []byte("data"),
	}
	_, err = server.StoreData(ctx1, &pb.StoreDataRequest{Record: record})
	require.NoError(t, err)

	registerReq2 := &pb.RegisterRequest{
		Login:             "user2",
		EncryptedPassword: []byte("pass"),
	}
	registerResp2, err := server.Register(context.Background(), registerReq2)
	require.NoError(t, err)

	userID2, err := auth.ParseToken(*server.cfg, registerResp2.AccessToken)
	require.NoError(t, err)

	ctx2 := auth.WithUserID(context.Background(), userID2.UserID)

	resp, err := server.DeleteData(ctx2, &pb.DeleteDataRequest{Id: "record-1"})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "Data not found or access denied")
	assert.Nil(t, resp)
}
