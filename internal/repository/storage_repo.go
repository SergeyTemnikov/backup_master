package repository

import (
	"backup_master/internal/model"
	"database/sql"
	"os"
	"path/filepath"
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

func (r *StorageRepository) CalcDirSize(root string) (int64, error) {
	var size int64

	err := filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return nil
	})

	return size, err
}

func (r *StorageRepository) GetUsedBytes(rootPath string) (int64, error) {
	var total int64

	err := filepath.Walk(rootPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})

	return total, err
}
