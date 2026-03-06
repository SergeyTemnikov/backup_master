package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func SendNotification(app fyne.App, title, message string) {
	app.SendNotification(&fyne.Notification{
		Title:   title,
		Content: message,
	})
}

func copyDir(src, dst string) error {
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source: %w", err)
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destination: %w", err)
	}

	if strings.HasPrefix(dstAbs, srcAbs+string(filepath.Separator)) {
		return fmt.Errorf("destination directory %q is inside source directory %q", dst, src)
	}

	if srcAbs == dstAbs {
		return fmt.Errorf("source and destination are the same directory: %q", src)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

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

func ParseCronToUI(
	cron string,
	periodSelect, minuteSelect, hourSelect, weekdaySelect, dayOfMonthSelect *widget.Select,
	refreshCB func(),
) {
	parts := strings.Fields(cron)
	if len(parts) < 6 {
		periodSelect.SetSelected("Каждый день")
		refreshCB()
		return
	}

	minute := parts[1]
	hour := parts[2]
	dayOfMonth := parts[3]
	weekday := parts[5]

	switch {
	case dayOfMonth == "*" && weekday == "*":
		if hour == "*" {
			periodSelect.SetSelected("Каждый час")
		} else {
			periodSelect.SetSelected("Каждый день")
		}

	case dayOfMonth == "*" && weekday != "*":
		periodSelect.SetSelected("Каждую неделю")

	case dayOfMonth != "*" && weekday == "*":
		periodSelect.SetSelected("Каждый месяц")

	default:
		periodSelect.SetSelected("Каждый день")
	}

	if minute != "*" {
		minuteSelect.SetSelected(fmt.Sprintf("%02d", parseCronInt(minute)))
	}
	if hour != "*" {
		hourSelect.SetSelected(fmt.Sprintf("%02d", parseCronInt(hour)))
	}
	if weekday != "*" {
		weekdayMap := map[string]string{
			"1": "Пн", "2": "Вт", "3": "Ср",
			"4": "Чт", "5": "Пт", "6": "Сб", "0": "Вс", "7": "Вс",
		}
		weekdaySelect.SetSelected(weekdayMap[weekday])
	}
	if dayOfMonth != "*" {
		dayOfMonthSelect.SetSelected(fmt.Sprintf("%02d", parseCronInt(dayOfMonth)))
	}

	refreshCB()
}

func parseCronInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func dirChecksum(path string) (string, error) {
	hasher := sha256.New()

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(hasher, f); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
