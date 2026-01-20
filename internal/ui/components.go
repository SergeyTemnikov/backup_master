package ui

import (
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
