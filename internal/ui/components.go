package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func Title(text string) fyne.CanvasObject {
	return widget.NewLabelWithStyle(
		text,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
}

func ThemeToggleButton() fyne.CanvasObject {
	app := fyne.CurrentApp()
	btn := widget.NewButton("", nil)

	current := "system"

	update := func() {
		switch current {
		case "dark":
			app.Settings().SetTheme(theme.DarkTheme())
			btn.SetIcon(theme.VisibilityOffIcon()) // 🌙
		case "light":
			app.Settings().SetTheme(theme.LightTheme())
			btn.SetIcon(theme.VisibilityIcon()) // ☀️
		default:
			app.Settings().SetTheme(nil)
			btn.SetIcon(theme.ComputerIcon())
		}
	}

	btn.OnTapped = func() {
		switch current {
		case "system":
			current = "dark"
		case "dark":
			current = "light"
		default:
			current = "system"
		}
		update()
	}

	update()
	return btn
}

func applyTheme(app fyne.App, mode string) {
	switch mode {
	case "dark":
		app.Settings().SetTheme(theme.DarkTheme())
	case "light":
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
