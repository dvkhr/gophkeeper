package repository

import (
	"database/sql"
	"time"

	"github.com/dvkhr/gophkeeper/pb"
)

// PostgresRepository — объединённая реализация всех репозиториев.
// Использует одно соединение с базой данных для работы с пользователями, данными и токенами.
type PostgresRepository struct {
	userRepo  *PostgresUserRepository
	dataRepo  *PostgresDataRepository
	tokenRepo *PostgresTokenRepository
}

// NewPostgresRepository создаёт новый экземпляр Repository с доступом к PostgreSQL.
// Все подсистемы (пользователи, данные, токены) используют одно соединение с БД.
func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{
		userRepo:  &PostgresUserRepository{db: db},
		dataRepo:  &PostgresDataRepository{db: db},
		tokenRepo: &PostgresTokenRepository{db: db},
	}
}

func (r *PostgresRepository) CreateUser(login, passwordHash string) (string, error) {
	return r.userRepo.CreateUser(login, passwordHash)
}

func (r *PostgresRepository) GetUserByLogin(login string) (*User, error) {
	return r.userRepo.GetUserByLogin(login)
}

func (r *PostgresRepository) SaveData(userID string, data *pb.DataRecord) error {
	return r.dataRepo.SaveData(userID, data)
}

func (r *PostgresRepository) GetAllData(userID string) ([]*pb.DataRecord, error) {
	return r.dataRepo.GetAllData(userID)
}

func (r *PostgresRepository) SaveRefreshToken(token, userID string, expiresAt time.Time) error {
	return r.tokenRepo.SaveRefreshToken(token, userID, expiresAt)
}

func (r *PostgresRepository) IsRefreshTokenRevoked(token string) (bool, error) {
	return r.tokenRepo.IsRefreshTokenRevoked(token)
}

func (r *PostgresRepository) RevokeRefreshToken(token string) error {
	return r.tokenRepo.RevokeRefreshToken(token)
}

func (r *PostgresRepository) DataExistsForUser(id string, userID string) (bool, error) {
	return r.dataRepo.DataExistsForUser(id, userID)
}

func (r *PostgresRepository) MarkDataAsDeleted(id string) error {
	return r.dataRepo.MarkDataAsDeleted(id)
}

func (r *PostgresRepository) GetUserIDByRefreshToken(token string) (string, error) {
	return r.tokenRepo.GetUserIDByRefreshToken(token)
}
