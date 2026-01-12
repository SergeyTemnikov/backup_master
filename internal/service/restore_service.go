package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type RestoreMode int

const (
	RestoreToNewFolder RestoreMode = iota
	RestoreOverwrite
)

type RestoreService struct{}

func NewRestoreService() *RestoreService {
	return &RestoreService{}
}

func (r *RestoreService) RestoreFile(
	backupPath string,
	targetDir string,
	overwrite bool,
) error {

	src, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return fmt.Errorf("backup file is empty: %s", backupPath)
	}

	_, err = src.Seek(0, 0)
	if err != nil {
		return err
	}

	backupName := filepath.Base(backupPath)
	originalName := restoreOriginalName(backupName)

	var dstPath string

	if overwrite {
		// перезапись рядом с backup
		dstPath = filepath.Join(
			filepath.Dir(backupPath),
			originalName,
		)
	} else {
		if targetDir == "" {
			return fmt.Errorf("target directory is required")
		}
		dstPath = filepath.Join(targetDir, originalName)
	}

	// Проверка перезаписи
	if _, err := os.Stat(dstPath); err == nil && !overwrite {
		return fmt.Errorf("file already exists: %s", dstPath)
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return err
	}

	if err := dst.Sync(); err != nil {
		return err
	}

	if written == 0 {
		return fmt.Errorf("restored file is empty after copy")
	}

	return nil
}

func restoreOriginalName(backupName string) string {
	// 1. Убираем .bak
	name := strings.TrimSuffix(backupName, ".bak")

	// 2. Отрезаем .YYYYMMDD_HHMMSS
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[:idx]
	}

	return name
}

func (r *RestoreService) RestoreFolder(
	backupDir string,
	targetRoot string,
	mode RestoreMode,
) error {

	info, err := os.Stat(backupDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("backup is not a directory")
	}

	backupName := filepath.Base(backupDir)
	originalName := restoreOriginalFolderName(backupName)

	var targetPath string

	switch mode {

	case RestoreToNewFolder:
		targetPath = filepath.Join(targetRoot, originalName)

		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("папка уже существует: %s", targetPath)
		}

	case RestoreOverwrite:
		targetPath = filepath.Join(targetRoot, originalName)

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			// если папки нет — просто создаём
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		}
	}

	return copyDirWithOverwrite(backupDir, targetPath)
}

func restoreOriginalFolderName(name string) string {
	// убираем _YYYYMMDD_HHMMSS или .YYYYMMDD_HHMMSS
	if idx := strings.LastIndexAny(name, "."); idx != -1 {
		name = name[:idx]
	}

	return name
}
