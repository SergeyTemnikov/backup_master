package model

import "time"

// Backup — результат выполнения задачи
type Backup struct {
	ID           int64     // primary key
	TaskID       int64     // FK -> tasks
	Status       string    // OK / ERROR
	SizeBytes    int64     // Размер копии
	ErrorMessage *string   // Текст ошибки (если была)
	StartedAt    time.Time // Время начала
	FinishedAt   time.Time // Время окончания
}
