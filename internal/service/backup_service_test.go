package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupFile_Success(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcPath, []byte("hello"), 0644)

	svc := NewBackupService()
	size, checksum, finalPath, err := svc.BackupFile(srcPath, dstDir)

	require.NoError(t, err)
	assert.Equal(t, int64(5), size)
	assert.NotEmpty(t, checksum)
	assert.FileExists(t, finalPath)
	assert.Contains(t, finalPath, ".bak")
}

func TestBackupFile_SourceNotFound(t *testing.T) {
	svc := NewBackupService()
	_, _, _, err := svc.BackupFile("/nonexistent/file.txt", t.TempDir())
	assert.Error(t, err)
}

func TestBackupFolder_Success(t *testing.T) {
	srcDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("bbb"), 0644)

	svc := NewBackupService()
	size, checksum, backupPath, err := svc.BackupFolder(srcDir, t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, int64(6), size) // aaa + bbb
	assert.NotEmpty(t, checksum)
	assert.DirExists(t, backupPath)
}
