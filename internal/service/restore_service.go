package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type RestoreService struct{}

func NewRestoreService() *RestoreService {
	return &RestoreService{}
}

func (r *RestoreService) RestoreFile(backupPath, targetDir string) error {
	// Открываем файл бэкапа
	src, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer src.Close()

	// 🔍 Проверяем размер бэкапа
	info, err := src.Stat()
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return fmt.Errorf("backup file is empty: %s", backupPath)
	}

	// ⚠️ На случай если файл уже читался ранее
	_, err = src.Seek(0, 0)
	if err != nil {
		return err
	}

	// Восстанавливаем имя
	backupName := filepath.Base(backupPath)
	originalName := restoreOriginalName(backupName)

	dstPath := filepath.Join(targetDir, originalName)

	// Создаём файл назначения
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 🔥 Копируем данные
	written, err := io.Copy(dst, src)
	if err != nil {
		return err
	}

	// 💾 Гарантируем запись на диск
	err = dst.Sync()
	if err != nil {
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
