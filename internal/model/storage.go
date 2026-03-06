package model

import "time"

type Storage struct {
	ID        int64
	Name      string
	Path      string
	MaxBytes  int64
	UsedBytes int64
	CreatedAt time.Time
}
