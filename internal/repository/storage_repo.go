package repository

import (
	"backup_master/internal/model"
	"database/sql"
)

type StorageRepository struct {
	db *sql.DB
}

func NewStorageRepository(db *sql.DB) *StorageRepository {
	return &StorageRepository{db: db}
}

// Все хранилища
func (r *StorageRepository) GetAll() ([]model.Storage, error) {
	rows, err := r.db.Query(`
		SELECT id, name, path, max_bytes, used_bytes, created_at
		FROM storages
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var storages []model.Storage
	for rows.Next() {
		var s model.Storage
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Path,
			&s.MaxBytes,
			&s.UsedBytes,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		storages = append(storages, s)
	}
	return storages, nil
}
