package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dvkhr/gophkeeper/pb"
)

var _ DataRepository = (*PostgresDataRepository)(nil)

type DataRepository interface {
	SaveData(userID string, data *pb.DataRecord) error
	GetAllData(userID string) ([]*pb.DataRecord, error)
}

type PostgresDataRepository struct {
	db *sql.DB
}

func NewDataRepository(db *sql.DB) DataRepository {
	return &PostgresDataRepository{db: db}
}

func (r *PostgresDataRepository) SaveData(userID string, data *pb.DataRecord) error {
	_, err := r.db.ExecContext(context.Background(),
		`INSERT INTO user_data (id, user_id, type, encrypted_data, metadata)
         VALUES ($1, $2, $3, $4, $5)
         ON CONFLICT (id) DO UPDATE SET
             type = EXCLUDED.type,
             encrypted_data = EXCLUDED.encrypted_data,
             metadata = EXCLUDED.metadata,
             updated_at = NOW()`,
		data.Id, userID, data.Type, data.EncryptedData, data.Metadata)

	if err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}
	return nil
}

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
			metadataRaw []byte // <-- читаем как []byte
		)

		if err := rows.Scan(
			&record.Id,
			&record.Type,
			&record.EncryptedData,
			&metadataRaw, // <-- сначала считываем как []byte
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
