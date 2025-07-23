package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dvkhr/gophkeeper/pb"
)

var _ DataRepository = (*PostgresDataRepository)(nil)

// DataRepository — интерфейс для работы с данными пользователя в базе данных.
type DataRepository interface {
	// SaveData сохраняет или обновляет запись пользователя в базе данных.
	SaveData(userID string, data *pb.DataRecord) error

	// GetAllData возвращает все неудалённые данные пользователя.
	// Данные возвращаются в порядке убывания времени обновления.
	GetAllData(userID string) ([]*pb.DataRecord, error)

	// DataExistsForUser проверяет, принадлежит ли запись пользователю
	DataExistsForUser(id, userID string) (bool, error)

	// MarkDataAsDeleted помечает запись как удаленную
	MarkDataAsDeleted(id string) error
}

// PostgresDataRepository — реализация DataRepository для PostgreSQL.
type PostgresDataRepository struct {
	db *sql.DB
}

// NewDataRepository создаёт новый экземпляр DataRepository.
func NewDataRepository(db *sql.DB) DataRepository {
	return &PostgresDataRepository{db: db}
}

// SaveData сохраняет или обновляет запись пользователя в базе данных.
func (r *PostgresDataRepository) SaveData(userID string, data *pb.DataRecord) error {
	_, err := r.db.ExecContext(context.Background(),
		`INSERT INTO user_data (id, user_id, type, encrypted_data, metadata)
         VALUES ($1, $2, $3, $4, $5)
         ON CONFLICT (id) DO UPDATE SET
    		 user_id = EXCLUDED.user_id,
             type = EXCLUDED.type,
             encrypted_data = EXCLUDED.encrypted_data,
             metadata = EXCLUDED.metadata,
             deleted = FALSE,
             updated_at = NOW()`,
		data.Id, userID, data.Type, data.EncryptedData, data.Metadata)

	if err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}
	return nil
}

// GetAllData возвращает все не удалённые данные пользователя из базы данных.
func (r *PostgresDataRepository) GetAllData(userID string) ([]*pb.DataRecord, error) {
	rows, err := r.db.QueryContext(context.Background(),
		`SELECT id, type, encrypted_data, metadata, EXTRACT(EPOCH FROM updated_at)::int
         FROM user_data
         WHERE user_id = $1 AND deleted = false
         ORDER BY updated_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all data: %w", err)
	}
	defer rows.Close()

	var records []*pb.DataRecord
	for rows.Next() {
		var (
			record      pb.DataRecord
			metadataRaw []byte
		)

		if err := rows.Scan(
			&record.Id,
			&record.Type,
			&record.EncryptedData,
			&metadataRaw,
			&record.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Теперь конвертируем JSON в map[string]string
		if len(metadataRaw) > 0 && string(metadataRaw) != "null" {
			err = json.Unmarshal(metadataRaw, &record.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		records = append(records, &record)
	}

	return records, nil
}

func (r *PostgresDataRepository) DataExistsForUser(id, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM user_data WHERE id = $1 AND user_id = $2)`,
		id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check data ownership: %w", err)
	}
	return exists, nil
}

func (r *PostgresDataRepository) MarkDataAsDeleted(id string) error {
	_, err := r.db.ExecContext(context.Background(),
		`UPDATE user_data SET deleted = TRUE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to mark data as deleted: %w", err)
	}
	return nil
}
