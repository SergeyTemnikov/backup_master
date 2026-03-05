package repository

import (
	"backup_master/internal/model"
	"database/sql"
	"strings"
	"time"
)

type BackupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

func (r *BackupRepository) Create(b *model.Backup) error {
	_, err := r.db.Exec(`
		INSERT INTO backups (
			task_id,
			status,
			size_bytes,
			error_message,
			checksum,
			started_at,
			finished_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		b.TaskID,
		b.Status,
		b.SizeBytes,
		b.ErrorMessage,
		b.Checksum,
		b.StartedAt,
		b.FinishedAt,
	)

	return err
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
		SELECT id, task_id, status, size_bytes, error_message, started_at, finished_at, checksum
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
			&b.Checksum,
		)
		if err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}

func (r *BackupRepository) GetByID(id int64) (*model.Backup, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, status, size_bytes, error_message, started_at, finished_at, checksum
		FROM backups
		ORDER BY finished_at DESC
		WHERE id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backup *model.Backup
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
			&b.Checksum,
		)
		if err != nil {
			return nil, err
		}
	}
	return backup, nil
}

func (r *BackupRepository) GetAll() ([]model.Backup, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, status, size_bytes, error_message, started_at, finished_at, checksum
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
			&b.Checksum,
		); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}

func (r *BackupRepository) GetHistory(limit int, statusFilter string, dateFilter *time.Time) ([]model.BackupWithTask, error) {
	query := `
        SELECT
            b.id,
            b.task_id,
            COALESCE(t.name, '—') AS task_name,
            COALESCE(t.source_path, '-') AS task_path,
            COALESCE(t.source_type, '-') AS task_type,
            b.status,
            b.size_bytes,
            b.error_message,
            b.started_at,
            b.finished_at,
			b.checksum
        FROM backups b
        LEFT JOIN tasks t ON t.id = b.task_id
    `

	conditions := []string{}
	args := []interface{}{}

	if statusFilter != "" {
		conditions = append(conditions, "b.status = ?")
		args = append(args, statusFilter)
	}

	if dateFilter != nil {
		conditions = append(conditions, "DATE(b.started_at) = DATE(?)")
		args = append(args, dateFilter.Format("2006-01-02"))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY b.started_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.BackupWithTask

	for rows.Next() {
		var b model.BackupWithTask
		if err := rows.Scan(
			&b.ID,
			&b.TaskID,
			&b.TaskName,
			&b.BackupPath,
			&b.SourceType,
			&b.Status,
			&b.SizeBytes,
			&b.ErrorMsg,
			&b.StartedAt,
			&b.FinishedAt,
			&b.Checksum,
		); err != nil {
			return nil, err
		}
		res = append(res, b)
	}

	return res, nil
}
