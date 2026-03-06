package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCron_EveryHour(t *testing.T) {
	cron, err := BuildCron("Каждый час", "30", "*", "*", "*")
	require.NoError(t, err)
	assert.Equal(t, "0 30 * * * *", cron)
}

func TestRestoreOriginalName(t *testing.T) {
	tests := []struct {
		backupName string
		want       string
	}{
		{"report.txt.20231001_120000.bak", "report.txt"},
		{"data.20231001_120000.bak", "data"},
		{"no_timestamp.bak", "no_timestamp"}, // без таймстампа
	}

	for _, tt := range tests {
		t.Run(tt.backupName, func(t *testing.T) {
			got := restoreOriginalName(tt.backupName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsTimestamp_Valid(t *testing.T) {
	assert.True(t, isTimestamp("20231001_120000"))
	assert.False(t, isTimestamp("not-a-timestamp"))
	assert.False(t, isTimestamp("20231001_12000"))
}
