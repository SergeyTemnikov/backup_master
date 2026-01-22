package repository

import (
	"backup_master/internal/model"
	"database/sql"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get() (*model.AppSettings, error) {
	row := r.db.QueryRow(`
		SELECT backup_root_path, max_storage_bytes, theme_mode
		FROM settings
		WHERE id = 1
	`)

	var s model.AppSettings
	err := row.Scan(
		&s.BackupRootPath,
		&s.MaxStorageBytes,
		&s.ThemeMode,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *SettingsRepository) Save(s *model.AppSettings) error {
	_, err := r.db.Exec(`
		UPDATE settings
		SET
			backup_root_path = ?,
			max_storage_bytes = ?,
			theme_mode = ?
		WHERE id = 1
	`,
		s.BackupRootPath,
		s.MaxStorageBytes,
		s.ThemeMode,
	)

	return err
}

func (r *SettingsRepository) UpdateBackupSettings(path string, maxBytes int64) error {
	_, err := r.db.Exec(`
		UPDATE settings
		SET backup_root_path = ?, max_storage_bytes = ?
		WHERE id = 1
	`, path, maxBytes)

	return err
}

func (r *SettingsRepository) UpdateTheme(mode string) error {
	_, err := r.db.Exec(`
		UPDATE settings
		SET theme_mode = ?
		WHERE id = 1
	`, mode)

	return err
}
