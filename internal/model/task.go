package model

import "time"

// Task — план резервного копирования
type Task struct {
	ID         int64     // primary key
	Name       string    // Название задачи
	SourcePath string    // Что копируем
	StorageID  int64     // Куда копируем (FK -> storages)
	Schedule   string    // Cron / описание расписания
	Enabled    bool      // Включена ли задача
	CreatedAt  time.Time // Дата создания
}
