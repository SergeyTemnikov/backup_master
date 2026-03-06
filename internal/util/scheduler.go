package util

import (
	"time"

	"github.com/robfig/cron/v3"
)

func CalculateNextRun(schedule string, from time.Time) (time.Time, error) {
	if schedule == "" {
		return time.Time{}, nil
	}

	c, err := cron.ParseStandard(schedule)
	if err != nil {
		return time.Time{}, err
	}

	return c.Next(from), nil
}

func UpdateNextRun(schedule string, from time.Time) *time.Time {
	next, err := CalculateNextRun(schedule, from)
	if err != nil {
		return nil
	}
	return &next
}
