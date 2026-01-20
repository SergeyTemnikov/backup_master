package repository

import (
	"backup_master/internal/model"
	"database/sql"
)

type BackupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

func (r *BackupRepository) CountAll() (int, error) {
	row := r.db.QueryRow(`SELECT COUNT(*) FROM backups`)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (r *BackupRepository) CountByStatus(status string) (int, error) {
	row := r.db.QueryRow(
		`SELECT COUNT(*) FROM backups WHERE status = ?`,
		status,
	)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (r *BackupRepository) GetLast(limit int) ([]model.Backup, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, status, size_bytes, error_message, started_at, finished_at
		FROM backups
		ORDER BY finished_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []model.Backup
	for rows.Next() {
		var b model.Backup
		err := rows.Scan(
			&b.ID,
			&b.TaskID,
			&b.Status,
			&b.SizeBytes,
			&b.ErrorMessage,
			&b.StartedAt,
			&b.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}

func (r *BackupRepository) GetAll() ([]model.Backup, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, status, size_bytes, error_message, started_at, finished_at
		FROM backups
		ORDER BY finished_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []model.Backup
	for rows.Next() {
		var b model.Backup
		if err := rows.Scan(
			&b.ID,
			&b.TaskID,
			&b.Status,
			&b.SizeBytes,
			&b.ErrorMessage,
			&b.StartedAt,
			&b.FinishedAt,
		); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}
