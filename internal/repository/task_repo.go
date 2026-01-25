package repository

import (
	"backup_master/internal/model"
	"database/sql"
	"time"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Все задачи
func (r *TaskRepository) GetAll() ([]model.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, name, source_path, source_type,schedule, enabled, created_at
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

// Ближайшие задачи (MVP — просто включённые)
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

// Количество ближайших (используется для dashboard)
func (r *TaskRepository) CountUpcoming(from, to time.Time) (int, error) {
	row := r.db.QueryRow(`
		SELECT COUNT(*) FROM tasks WHERE enabled = 1
	`)
	var count int
	err := row.Scan(&count)
	return count, err
}

// Создание задачи
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

// Удаление задачи
func (r *TaskRepository) Delete(taskID int64) error {
	_, err := r.db.Exec(`DELETE FROM tasks WHERE id = ?`, taskID)
	return err
}

// Включить / выключить
func (r *TaskRepository) SetEnabled(taskID int64, enabled bool) error {
	_, err := r.db.Exec(`
		UPDATE tasks SET enabled = ? WHERE id = ?
	`, enabled, taskID)
	return err
}

// Только включённые задачи
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
