package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
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
