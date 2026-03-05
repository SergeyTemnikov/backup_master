package model

import "time"

type Task struct {
	ID         int64     // primary key
	Name       string    // Название задачи
	Schedule   string    // Cron / описание расписания
	SourcePath string    // Что копируем
	SourceType string    // "file" || "folder"
	Enabled    bool      // Включена ли задача
	CreatedAt  time.Time // Дата создания
}
