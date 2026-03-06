package service

import (
	"backup_master/internal/model"
	"backup_master/internal/repository/mocks"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAppService_CheckStorageLimit_Exceeded(t *testing.T) {
	backupDir := t.TempDir()

	largeFile := filepath.Join(backupDir, "large.dat")
	err := os.WriteFile(largeFile, make([]byte, 200), 0644)
	require.NoError(t, err)

	mockSettingsRepo := new(mocks.SettingsRepositoryInterface)

	svc := &AppService{
		SettingsRepo: mockSettingsRepo,
		Settings: &model.AppSettings{
			BackupRootPath:  backupDir,
			MaxStorageBytes: 100,
		},
	}

	err = svc.CheckStorageLimit()

	require.Error(t, err, "Ожидалась ошибка о превышении лимита")

	assert.Contains(t, err.Error(), "превышен лимит")

	mockSettingsRepo.AssertExpectations(t) // ← Убери, если Get() не вызывался
}

// app_service_test.go

func TestAppService_RunManualBackup_SaveMetadata(t *testing.T) {
	mockBackupRepo := new(mocks.BackupRepositoryInterface)

	mockBackupRepo.On("Create", mock.MatchedBy(func(b *model.Backup) bool {
		return b.Status == "OK" &&
			b.SourceType == "file" &&
			b.SizeBytes > 0 &&
			b.Checksum != "" &&
			b.TargetPath != ""
	})).Return(nil)

	backupDir := t.TempDir()
	svc := &AppService{
		BackupRepo: mockBackupRepo,
		Settings: &model.AppSettings{
			BackupRootPath: backupDir,
		},
		BackupSvc: NewBackupService(),
	}

	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "file.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	require.NoError(t, err)

	dstDir := t.TempDir()

	err = svc.RunManualBackup(srcFile, dstDir)

	assert.NoError(t, err)

	mockBackupRepo.AssertExpectations(t)
}
