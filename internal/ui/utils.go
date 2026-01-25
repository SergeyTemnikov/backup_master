package ui

import (
	"backup_master/internal/model"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/robfig/cron/v3"
)

func bytesToGB(b int64) float64 {
	return float64(b) / (1024 * 1024 * 1024)
}

func gbToBytes(gb float64) int64 {
	return int64(gb * 1024 * 1024 * 1024)
}

func applyTheme(app fyne.App, mode string) {
	switch mode {
	case "Темная":
		app.Settings().SetTheme(theme.DarkTheme())
	case "Светлая":
		app.Settings().SetTheme(theme.LightTheme())
	default:
		app.Settings().SetTheme(nil) // system
	}
}
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div),
		"KMGTPE"[exp],
	)
}

func HumanizeCron(cron string) string {
	parts := strings.Fields(cron)
	if len(parts) != 6 {
		return cron // fallback
	}

	minute := parts[1]
	hour := parts[2]
	dayOfMonth := parts[3]
	month := parts[4]
	weekday := parts[5]

	// Каждый час
	if hour == "*" && dayOfMonth == "*" && month == "*" && weekday == "*" {
		return fmt.Sprintf(
			"В %s минут каждого часа",
			minute,
		)
	}

	// Каждый день
	if dayOfMonth == "*" && month == "*" && weekday == "*" {
		return fmt.Sprintf(
			"Каждый день в %s:%s",
			pad(hour), pad(minute),
		)
	}

	// Каждую неделю
	if dayOfMonth == "*" && month == "*" {
		return fmt.Sprintf(
			"Каждый %s в %s:%s",
			weekdayName(weekday),
			pad(hour), pad(minute),
		)
	}

	// Каждый месяц
	if month == "*" {
		return fmt.Sprintf(
			"Каждого %s числа в %s:%s",
			dayOfMonth,
			pad(hour), pad(minute),
		)
	}

	return cron
}

func pad(v string) string {
	if len(v) == 1 {
		return "0" + v
	}
	return v
}

func weekdayName(v string) string {
	switch v {
	case "0":
		return "воскресенье"
	case "1":
		return "понедельник"
	case "2":
		return "вторник"
	case "3":
		return "среду"
	case "4":
		return "четверг"
	case "5":
		return "пятницу"
	case "6":
		return "субботу"
	default:
		return v
	}
}

func NextRunAt(cronExpr string, from time.Time) (time.Time, error) {
	parser := cron.NewParser(
		cron.Second |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow,
	)

	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(from), nil
}

func NextRunString(task model.Task) string {
	next, err := NextRunAt(task.Schedule, time.Now())
	if err != nil {
		return "ошибка расписания"
	}

	return next.Format("02.01 15:04")
}
