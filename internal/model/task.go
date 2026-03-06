package model

import "time"

type Task struct {
	ID         int64
	Name       string
	Schedule   string
	SourcePath string
	SourceType string
	Enabled    bool
	CreatedAt  time.Time
}
