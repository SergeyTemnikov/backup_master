package ui

import (
	"backup_master/internal/service"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowFirstRunDialog(svc *service.AppService, w fyne.Window) {
	var (
		selectedPath string
		maxGB        float64 = 50
		themeMode            = "Системная"
	)

	// ---- UI elements ----

	title := widget.NewLabelWithStyle(
		"Первичная настройка",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	pathLabel := widget.NewLabel("Папка не выбрана")

	selectFolderBtn := widget.NewButton("Выбрать папку", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri == nil || err != nil {
				return
			}

			selectedPath = uri.Path()
			pathLabel.SetText(selectedPath)
		}, w)
	})

	limitLabel := widget.NewLabel(fmt.Sprintf("Лимит: %.0f ГБ", maxGB))

	limitSlider := widget.NewSlider(1, 1000)
	limitSlider.Value = maxGB
	limitSlider.Step = 1
	limitSlider.OnChanged = func(v float64) {
		maxGB = v
		limitLabel.SetText(fmt.Sprintf("Лимит: %.0f ГБ", v))
	}

	themeRadio := widget.NewRadioGroup(
		[]string{"Системная", "Светлая", "Темная"},
		func(v string) {
			themeMode = v
		},
	)
	themeRadio.SetSelected("Системная")

	// ---- Dialog (will be assigned later) ----
	var dlg *dialog.CustomDialog

	saveBtn := widget.NewButton("Сохранить", func() {
		if selectedPath == "" {
			dialog.ShowInformation(
				"Ошибка",
				"Пожалуйста, выберите папку для хранения бэкапов",
				w,
			)
			return
		}

		svc.Settings.BackupRootPath = selectedPath
		svc.Settings.MaxStorageBytes = int64(maxGB * 1024 * 1024 * 1024)
		svc.Settings.ThemeMode = themeMode

		if err := svc.SettingsRepo.Save(svc.Settings); err != nil {
			dialog.ShowError(err, w)
			return
		}

		applyTheme(fyne.CurrentApp(), themeMode)

		dlg.Hide() // ✅ закрываем только при успехе
	})

	exitBtn := widget.NewButton("Выход", func() {
		w.Close()
	})

	// ---- Content ----

	content := container.NewVBox(
		title,
		widget.NewSeparator(),

		widget.NewLabel("Папка для хранения бэкапов"),
		pathLabel,
		selectFolderBtn,

		widget.NewSeparator(),

		widget.NewLabel("Лимит хранилища"),
		limitSlider,
		limitLabel,

		widget.NewSeparator(),

		widget.NewLabel("Тема интерфейса"),
		themeRadio,

		widget.NewSeparator(),
	)

	dlg = dialog.NewCustom(
		"Добро пожаловать в Backup Master",
		"",
		content,
		w,
	)

	btns := []fyne.CanvasObject{exitBtn, saveBtn}

	dlg.SetButtons(btns)

	dlg.Show()
}
