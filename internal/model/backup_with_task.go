package model

import "time"

type BackupWithTask struct {
	ID         int64
	TaskID     *int64
	TaskName   string
	BackupPath string
	SourceType string
	Status     string
	SizeBytes  int64
	ErrorMsg   *string
	StartedAt  time.Time
	FinishedAt time.Time
	Checksum   string
}
