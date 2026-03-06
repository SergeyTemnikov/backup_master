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
			source_path,
			source_type,
			target_path,
			error_message,
			checksum,
			started_at,
			finished_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		b.TaskID,
		b.Status,
		b.SizeBytes,
		b.SourcePath,
		b.SourceType,
		b.TargetPath,
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
		SELECT 
			id,
			task_id,
			status,
			size_bytes,
			source_path,
			source_type,
			target_path,
			error_message,
			started_at,
			finished_at,
			checksum
		FROM backups
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
			&b.SourcePath,
			&b.SourceType,
			&b.TargetPath,
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
	row := r.db.QueryRow(`
		SELECT 
			id,
			task_id,
			status,
			size_bytes,
			source_path,
			source_type,
			target_path,
			error_message,
			started_at,
			finished_at,
			checksum
		FROM backups
		WHERE id = ?
	`, id)

	backup := &model.Backup{}
	err := row.Scan(
		&backup.ID,
		&backup.TaskID,
		&backup.Status,
		&backup.SizeBytes,
		&backup.SourcePath,
		&backup.SourceType,
		&backup.TargetPath,
		&backup.ErrorMessage,
		&backup.StartedAt,
		&backup.FinishedAt,
		&backup.Checksum,
	)
	if err != nil {
		return nil, err
	}
	return backup, nil
}

func (r *BackupRepository) GetAll() ([]model.Backup, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, status, size_bytes, target_path, error_message, started_at, finished_at, checksum
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
			&b.TargetPath,
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

func (r *BackupRepository) GetHistory(limit int, statusFilter string, dateFilter *time.Time, id *int64) ([]model.BackupWithTask, error) {
	query := `
		SELECT
			b.id,
			b.task_id,
			COALESCE(t.name, 'Ручной бэкап') AS task_name,
			b.source_path,
			b.source_type,
			b.status,
			b.size_bytes,
			b.target_path,
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

	if id != nil {
		if *id == -1 {
			conditions = append(conditions, "task_name = 'Ручной бэкап'")
		} else {
			conditions = append(conditions, "b.task_id = ?")
			args = append(args, *id)
		}
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
			&b.SourcePath,
			&b.SourceType,
			&b.Status,
			&b.SizeBytes,
			&b.TargetPath,
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
