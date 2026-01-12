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

// BackupFile копирует файл srcPath в dstDir
func (b *BackupService) BackupFile(srcPath, dstDir string) (int64, error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return 0, fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat source: %w", err)
	}

	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("not a regular file")
	}

	dstPath := filepath.Join(
		dstDir,
		info.Name()+"."+time.Now().Format("20060102_150405")+".bak",
	)

	dst, err := os.Create(dstPath)
	if err != nil {
		return 0, fmt.Errorf("create destination: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return written, fmt.Errorf("copy: %w", err)
	}

	return written, nil
}

func (b *BackupService) BackupFolder(sourceDir, targetRoot string) (int64, error) {
	info, err := os.Stat(sourceDir)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 0, ErrNotDirectory
	}

	// Считаем размер ДО копирования
	size, err := dirSize(sourceDir)
	if err != nil {
		return 0, err
	}

	timestamp := time.Now().Format("20060102_150405")
	folderName := filepath.Base(sourceDir) + "." + timestamp
	targetDir := filepath.Join(targetRoot, folderName)

	if err := copyDir(sourceDir, targetDir); err != nil {
		return 0, err
	}

	return size, nil
}
