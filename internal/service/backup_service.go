package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type BackupService struct{}

func NewBackupService() *BackupService {
	return &BackupService{}
}

var ErrNotDirectory = errors.New("source path is not a directory")

func (b *BackupService) BackupFile(srcPath, dstDir string) (int64, string, string, error) {

	src, err := os.Open(srcPath)
	if err != nil {
		return 0, "", "", err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return 0, "", "", err
	}
	if !info.Mode().IsRegular() {
		return 0, "", "", fmt.Errorf("not a regular file")
	}

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return 0, "", "", err
	}

	finalName := info.Name() + "." + time.Now().Format("20060102_150405") + ".bak"
	finalPath := filepath.Join(dstDir, finalName)

	tmpFile, err := os.CreateTemp(dstDir, finalName+".*.tmp")
	if err != nil {
		return 0, "", "", err
	}

	tmpPath := tmpFile.Name()

	written, err := io.Copy(tmpFile, src)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return 0, "", "", err
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return 0, "", "", err
	}

	tmpFile.Close()

	checksum, err := fileChecksum(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return 0, "", "", err
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return 0, "", "", err
	}

	return written, checksum, finalPath, nil
}

func (b *BackupService) BackupFolder(sourceDir, targetRoot string) (int64, string, string, error) {

	info, err := os.Stat(sourceDir)
	if err != nil {
		return 0, "", "", err
	}

	if !info.IsDir() {
		return 0, "", "", ErrNotDirectory
	}

	size, err := dirSize(sourceDir)
	if err != nil {
		return 0, "", "", err
	}

	timestamp := time.Now().Format("20060102_150405")
	folderName := filepath.Base(sourceDir) + "." + timestamp
	targetDir := filepath.Join(targetRoot, folderName)

	if err := copyDir(sourceDir, targetDir); err != nil {
		return 0, "", "", err
	}

	checksum, err := dirChecksum(targetDir)
	if err != nil {
		return 0, "", "", err
	}

	return size, checksum, targetDir, nil
}
