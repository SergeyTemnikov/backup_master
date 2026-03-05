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

	if targetDir == "" {
		return fmt.Errorf("target directory is required")
	}

	srcInfo, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("stat backup: %w", err)
	}
	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("backup is not a regular file")
	}
	if srcInfo.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}

	backupName := filepath.Base(backupPath)
	originalName := restoreOriginalName(backupName)
	dstPath := filepath.Join(targetDir, originalName)
	tmpPath := dstPath + ".tmp"

	if _, err := os.Stat(dstPath); err == nil {
		if !overwrite {
			return fmt.Errorf("file already exists: %s", dstPath)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	backupChecksum, err := fileChecksum(backupPath)
	if err != nil {
		return fmt.Errorf("checksum backup: %w", err)
	}

	src, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer src.Close()

	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, src); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("copy: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sync: %w", err)
	}

	tmpFile.Close()

	tmpChecksum, err := fileChecksum(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("checksum restored: %w", err)
	}

	if backupChecksum != tmpChecksum {
		os.Remove(tmpPath)
		return fmt.Errorf("checksum mismatch: backup and restored file differ")
	}

	if err := os.Rename(tmpPath, dstPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

func (r *RestoreService) RestoreFileWithChecksum(
	backupPath string,
	targetDir string,
	expectedChecksum string,
	overwrite bool,
) error {

	if targetDir == "" {
		return fmt.Errorf("требуется папка назначения")
	}

	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("backup не найден: %w", err)
	}

	if backupInfo.Size() == 0 {
		return fmt.Errorf("backup файл пуст")
	}

	backupChecksum, err := fileChecksum(backupPath)
	if err != nil {
		return fmt.Errorf("ошибка вычисления checksum backup: %w", err)
	}

	if backupChecksum != expectedChecksum {
		return fmt.Errorf("backup повреждён: checksum mismatch")
	}

	originalName := restoreOriginalName(filepath.Base(backupPath))
	dstPath := filepath.Join(targetDir, originalName)
	tmpPath := dstPath + ".tmp"

	if _, err := os.Stat(dstPath); err == nil && !overwrite {
		return fmt.Errorf("файл уже существует")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	src, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer src.Close()

	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	if _, err := io.Copy(tmpFile, src); err != nil {
		return fmt.Errorf("ошибка копирования: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	restoredChecksum, err := fileChecksum(tmpPath)
	if err != nil {
		return err
	}

	if restoredChecksum != backupChecksum {
		return fmt.Errorf("restore verification failed: checksum mismatch")
	}

	if err := os.Rename(tmpPath, dstPath); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}

	dir, err := os.Open(targetDir)
	if err == nil {
		dir.Sync()
		dir.Close()
	}

	return nil
}

func restoreOriginalName(name string) string {
	name = strings.TrimSuffix(name, ".bak")

	if len(name) > 16 {
		suffix := name[len(name)-15:]
		if isTimestamp(suffix) {
			return name[:len(name)-16]
		}
	}

	return name
}

func (r *RestoreService) RestoreFolder(
	backupDir string,
	targetRoot string,
	mode RestoreMode,
) error {

	if targetRoot == "" {
		return fmt.Errorf("target root required")
	}

	info, err := os.Stat(backupDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("backup is not a directory")
	}

	originalName := restoreOriginalFolderName(filepath.Base(backupDir))
	finalPath := filepath.Join(targetRoot, originalName)
	tempPath := finalPath + ".tmp"

	if mode == RestoreToNewFolder {
		if _, err := os.Stat(finalPath); err == nil {
			return fmt.Errorf("folder exists")
		}
	}

	if err := os.MkdirAll(targetRoot, 0755); err != nil {
		return err
	}

	if err := copyDirWithOverwrite(backupDir, tempPath); err != nil {
		os.RemoveAll(tempPath)
		return err
	}

	if mode == RestoreOverwrite {
		os.RemoveAll(finalPath)
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		os.RemoveAll(tempPath)
		return err
	}

	return nil
}

func restoreOriginalFolderName(name string) string {
	if len(name) > 16 {
		suffix := name[len(name)-15:]
		if isTimestamp(suffix) {
			return name[:len(name)-16]
		}
	}
	return name
}

func isTimestamp(s string) bool {
	if len(s) != 15 {
		return false
	}
	return s[8] == '_'
}
