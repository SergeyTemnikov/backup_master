package model

import "time"

// Task — план резервного копирования
type Task struct {
	ID         int64     // primary key
	Name       string    // Название задачи
	SourcePath string    // Что копируем
	SourceType string    // "file" | "folder"
	Schedule   string    // Cron / описание расписания
	Enabled    bool      // Включена ли задача
	CreatedAt  time.Time // Дата создания
}
