package model

import "time"

type Backup struct {
	ID           int64
	TaskID       *int64
	Status       string
	SizeBytes    int64
	SourcePath   string
	SourceType   string
	TargetPath   string
	ErrorMessage *string
	StartedAt    time.Time
	FinishedAt   time.Time
	Checksum     string
}
