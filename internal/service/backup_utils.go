package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func copyDir(src, dst string) error {
	// Приводим пути к абсолютным и нормализуем их
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source: %w", err)
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destination: %w", err)
	}

	// Убедимся, что dst не является подкаталогом src
	// Добавляем разделитель, чтобы избежать ложных совпадений (например, /data и /data2)
	if strings.HasPrefix(dstAbs, srcAbs+string(filepath.Separator)) {
		return fmt.Errorf("destination directory %q is inside source directory %q", dst, src)
	}

	// Если src и dst — один и тот же путь (после нормализации)
	if srcAbs == dstAbs {
		return fmt.Errorf("source and destination are the same directory: %q", src)
	}

	// Читаем содержимое исходной директории
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Создаём целевую директорию
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Копируем каждый элемент
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func dirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}

func copyDirWithOverwrite(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}

func BuildCron(
	period string,
	minute string,
	hour string,
	weekday string,
	dayOfMonth string,
) (string, error) {

	switch period {

	case "Каждый час":
		return fmt.Sprintf("0 %s * * * *", minute), nil

	case "Каждый день":
		return fmt.Sprintf("0 %s %s * * *", minute, hour), nil

	case "Каждую неделю":
		wd := map[string]string{
			"Пн": "1", "Вт": "2", "Ср": "3",
			"Чт": "4", "Пт": "5", "Сб": "6", "Вс": "0",
		}[weekday]

		return fmt.Sprintf("0 %s %s * * %s", minute, hour, wd), nil

	case "Каждый месяц":
		return fmt.Sprintf("0 %s %s %s * *", minute, hour, dayOfMonth), nil
	}

	return "", fmt.Errorf("неизвестный период")
}
