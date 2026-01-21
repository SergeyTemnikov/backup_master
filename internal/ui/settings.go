package ui

import (
	"backup_master/internal/service"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewSettings(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	// ====== ТЕМА ======
	themeRadio := widget.NewRadioGroup(
		[]string{"Системная", "Светлая", "Темная"},
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

	// ====== ЛИМИТ ХРАНИЛИЩА ======
	diskBytes, err := service.GetDiskTotalBytes(svc.Settings.BackupRootPath)
	if err != nil {
		diskBytes = 500 * 1024 * 1024 * 1024 // fallback 500 GB
	}

	maxGB := bytesToGB(diskBytes)
	currentGB := bytesToGB(svc.Settings.MaxStorageBytes)

	limitLabel := widget.NewLabel("")
	limitSlider := widget.NewSlider(1, maxGB)
	limitSlider.Step = 1
	limitSlider.SetValue(currentGB)

	updateLimitLabel := func(v float64) {
		limitLabel.SetText(
			fmt.Sprintf("Лимит: %.0f GB из %.0f GB", v, maxGB),
		)
	}

	limitSlider.OnChanged = func(v float64) {
		updateLimitLabel(v)
	}

	limitSlider.OnChangeEnded = func(v float64) {
		svc.Settings.MaxStorageBytes = gbToBytes(v)
		_ = svc.SettingsRepo.Save(svc.Settings)
	}

	updateLimitLabel(currentGB)

	return container.NewVScroll(
		container.NewVBox(
			Title("Тема"),
			themeRadio,

			Title("Хранилище"),
			pathLabel,
			selectDir,
			limitLabel,
			limitSlider,
		),
	)
}
