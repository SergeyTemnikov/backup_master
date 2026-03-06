package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreFileWithChecksum_Success(t *testing.T) {
	srcDir := t.TempDir()
	backupDir := t.TempDir()
	restoreDir := t.TempDir()

	originalPath := filepath.Join(srcDir, "doc.txt")
	os.WriteFile(originalPath, []byte("important data"), 0644)

	backupSvc := NewBackupService()
	_, expectedChecksum, backupPath, _ := backupSvc.BackupFile(originalPath, backupDir)

	restoreSvc := NewRestoreService()
	err := restoreSvc.RestoreFileWithChecksum(backupPath, restoreDir, expectedChecksum, false)

	require.NoError(t, err)

	restoredPath := filepath.Join(restoreDir, "doc.txt")
	restored, _ := os.ReadFile(restoredPath)
	assert.Equal(t, "important data", string(restored))
}

func TestRestoreFileWithChecksum_WrongChecksum(t *testing.T) {
	backupDir := t.TempDir()
	restoreDir := t.TempDir()

	backupPath := filepath.Join(backupDir, "fake.bak")
	os.WriteFile(backupPath, []byte("corrupted"), 0644)

	restoreSvc := NewRestoreService()
	err := restoreSvc.RestoreFileWithChecksum(backupPath, restoreDir, "wrong_checksum", false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestRestoreFolder_OverwriteMode(t *testing.T) {
	backupRoot := t.TempDir()
	restoreRoot := t.TempDir()
	sourceV1 := filepath.Join(backupRoot, "source_v1")
	require.NoError(t, os.MkdirAll(sourceV1, 0755))
	os.WriteFile(filepath.Join(sourceV1, "file_a.txt"), []byte("content A v1"), 0644)
	os.WriteFile(filepath.Join(sourceV1, "file_b.txt"), []byte("content B v1"), 0644)
	os.MkdirAll(filepath.Join(sourceV1, "subdir"), 0755)
	os.WriteFile(filepath.Join(sourceV1, "subdir", "nested.txt"), []byte("nested v1"), 0644)

	backupSvc := NewBackupService()
	_, expectedChecksum, backupPath, err := backupSvc.BackupFolder(sourceV1, backupRoot)
	require.NoError(t, err)
	require.NotEmpty(t, backupPath)
	t.Logf("Backup created at: %s", backupPath)

	originalTarget := filepath.Join(restoreRoot, "source_v1")
	require.NoError(t, os.MkdirAll(originalTarget, 0755))
	os.WriteFile(filepath.Join(originalTarget, "file_a.txt"), []byte("content A v2 - SHOULD BE OVERWRITTEN"), 0644)
	os.WriteFile(filepath.Join(originalTarget, "file_old.txt"), []byte("this file should be DELETED"), 0644)

	restoreSvc := NewRestoreService()
	err = restoreSvc.RestoreFolderWithChecksum(
		backupPath,
		restoreRoot,
		expectedChecksum,
		RestoreOverwrite,
	)
	require.NoError(t, err)

	restoredPath := filepath.Join(restoreRoot, "source_v1")

	contentA, err := os.ReadFile(filepath.Join(restoredPath, "file_a.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content A v1", string(contentA), "file_a.txt должен быть перезаписан версией из бэкапа")

	contentB, err := os.ReadFile(filepath.Join(restoredPath, "file_b.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content B v1", string(contentB))

	nested, err := os.ReadFile(filepath.Join(restoredPath, "subdir", "nested.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested v1", string(nested))

	_, err = os.Stat(filepath.Join(restoredPath, "file_old.txt"))
	assert.True(t, os.IsNotExist(err), "file_old.txt должен быть удалён при перезаписи")

	restoredChecksum, err := dirChecksum(restoredPath)
	require.NoError(t, err)
	assert.Equal(t, expectedChecksum, restoredChecksum, "Хеш восстановленной папки должен совпадать с бэкапом")
}
