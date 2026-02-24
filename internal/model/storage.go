package model

import "time"

type Storage struct {
	ID        int64     // primary key
	Name      string    // Название (C:, D:, Backups)
	Path      string    // Путь к папке
	MaxBytes  int64     // Максимально разрешенный объём
	UsedBytes int64     // Занято
	CreatedAt time.Time // Дата добавления
}
