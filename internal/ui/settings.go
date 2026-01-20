package ui

import (
	"backup_master/internal/service"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewSettings(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	// ====== ТЕМА ======
	themeRadio := widget.NewRadioGroup(
		[]string{"system", "light", "dark"},
		func(value string) {
			svc.Settings.ThemeMode = value
			applyTheme(fyne.CurrentApp(), value)
			_ = svc.SettingsRepo.Save(svc.Settings)
		},
	)
	themeRadio.SetSelected(svc.Settings.ThemeMode)

	// ====== ПАПКА ======
	pathLabel := widget.NewLabel(svc.Settings.BackupRootPath)

	selectDir := widget.NewButton("Выбрать папку", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri == nil {
				return
			}
			path := uri.Path()
			svc.Settings.BackupRootPath = path
			pathLabel.SetText(path)
			_ = svc.SettingsRepo.Save(svc.Settings)
		}, w)
	})

	// ====== ЛИМИТ ======
	limitEntry := widget.NewEntry()
	limitEntry.SetText(fmt.Sprintf("%d", svc.Settings.MaxStorageBytes))

	saveLimit := widget.NewButton("Сохранить лимит", func() {
		if v, err := strconv.ParseInt(limitEntry.Text, 10, 64); err == nil {
			svc.Settings.MaxStorageBytes = v
			_ = svc.SettingsRepo.Save(svc.Settings)
		}
	})

	return container.NewVScroll(
		container.NewVBox(
			Title("Настройки"),
			Title("Тема"),
			themeRadio,
			layout.NewSpacer(),

			Title("Хранилище"),
			pathLabel,
			selectDir,
			widget.NewLabel("Максимальный размер (байты)"),
			limitEntry,
			saveLimit,
		),
	)
}
