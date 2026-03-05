package model

import "time"

type Backup struct {
	ID           int64     // primary key
	TaskID       *int64    // FK -> tasks
	Status       string    // OK / ERROR
	SizeBytes    int64     // Размер копии
	SourcePath   string    // Что копируем
	SourceType   string    // "file" || "folder"
	TargetPath   string    // Путь до копии
	ErrorMessage *string   // Текст ошибки (если была)
	StartedAt    time.Time // Время начала
	FinishedAt   time.Time // Время окончания
	Checksum     string    // Контрольная сумма
}
