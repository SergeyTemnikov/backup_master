package repository

import (
	"backup_master/internal/model"
	"backup_master/internal/util"
	"database/sql"
	"time"
)

type TaskRepositoryInterface interface {
	GetAll() ([]model.Task, error)
	GetUpcoming(limit int) ([]model.Task, error)
	CountUpcoming(from, to time.Time) (int, error)
	Create(task *model.Task) error
	Update(task *model.Task) error
	Delete(taskID int64) error
	SetEnabled(taskID int64, enabled bool) error
	GetEnabled() ([]model.Task, error)
}

var _ TaskRepositoryInterface = (*TaskRepository)(nil)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) GetAll() ([]model.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, name, source_path, source_type, schedule,  enabled, created_at
		FROM tasks
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.SourcePath,
			&t.SourceType,
			&t.Schedule,
			&t.Enabled,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) GetUpcoming(limit int) ([]model.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, name, source_path, source_type, schedule, enabled, created_at
		FROM tasks
		WHERE enabled = 1
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.SourcePath,
			&t.SourceType,
			&t.Schedule,
			&t.Enabled,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) CountUpcoming(from, to time.Time) (int, error) {
	rows, err := r.db.Query(`
		SELECT schedule FROM tasks WHERE enabled = 1
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var schedule string
		if err := rows.Scan(&schedule); err != nil {
			return 0, err
		}

		next, err := util.CalculateNextRun(schedule, from)
		if err != nil {
			continue
		}

		if !next.IsZero() && next.After(from) && next.Before(to) {
			count++
		}
	}

	return count, nil
}

func (r *TaskRepository) Create(task *model.Task) error {
	_, err := r.db.Exec(`
		INSERT INTO tasks (
			name,
			source_path,
			source_type,
			schedule,
			enabled,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		task.Name,
		task.SourcePath,
		task.SourceType,
		task.Schedule,
		task.Enabled,
		task.CreatedAt,
	)
	return err
}

func (r *TaskRepository) Update(task *model.Task) error {
	_, err := r.db.Exec(
		`
		UPDATE tasks 
		SET name=?, source_path=?, source_type=?, schedule=?, enabled=? 
		WHERE id=?`,
		task.Name,
		task.SourcePath,
		task.SourceType,
		task.Schedule,
		task.Enabled,
		task.ID,
	)
	return err
}

func (r *TaskRepository) Delete(taskID int64) error {
	_, err := r.db.Exec(`DELETE FROM tasks WHERE id = ?`, taskID)
	return err
}

func (r *TaskRepository) SetEnabled(taskID int64, enabled bool) error {
	_, err := r.db.Exec(`
		UPDATE tasks SET enabled = ? WHERE id = ?
	`, enabled, taskID)
	return err
}

func (r *TaskRepository) GetEnabled() ([]model.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, name, source_path, source_type, schedule, enabled, created_at
		FROM tasks
		WHERE enabled = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.SourcePath,
			&t.SourceType,
			&t.Schedule,
			&t.Enabled,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
