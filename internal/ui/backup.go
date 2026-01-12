package ui

import (
	"backup_master/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	backupMode   = "file" // file | folder
	fileURI      fyne.URI
	sourceFolder string
	targetFolder string
)

func NewBackup(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	// ===== ВЫБОР ФАЙЛА =====
	fileLabel := widget.NewLabel("Файл для копирования:")
	fileButton := widget.NewButton("Выбрать файл", func() {
		dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if f == nil {
				return
			}
			fileURI = f.URI()
			fileLabel.SetText("Файл: " + fileURI.Path())
		}, w)
	})

	fileBlock := container.NewVBox(fileLabel, fileButton)

	// ===== ВЫБОР ПАПКИ-ИСТОЧНИКА =====
	sourceLabel := widget.NewLabel("Папка для копирования:")
	sourceButton := widget.NewButton("Выбрать папку", func() {
		dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if f == nil {
				return
			}
			sourceFolder = f.Path()
			sourceLabel.SetText("Папка: " + sourceFolder)
		}, w)
	})

	sourceBlock := container.NewVBox(sourceLabel, sourceButton)
	sourceBlock.Hide()

	// ===== РЕЖИМ БЭКАПА =====
	modeSelector := widget.NewRadioGroup(
		[]string{"Файл", "Папка"},
		func(s string) {
			if s == "Файл" {
				backupMode = "file"

				fileBlock.Show()
				sourceBlock.Hide()
			} else {
				backupMode = "folder"

				fileBlock.Hide()
				sourceBlock.Show()
			}
		},
	)
	modeSelector.SetSelected("Файл")

	// ===== ВЫБОР ПАПКИ НАЗНАЧЕНИЯ =====
	targetLabel := widget.NewLabel("Папка назначения:")
	targetButton := widget.NewButton("Выбрать папку назначения", func() {
		dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if f == nil {
				return
			}
			targetFolder = f.Path()
			targetLabel.SetText("Назначение: " + targetFolder)
		}, w)
	})

	targetBlock := container.NewVBox(targetLabel, targetButton)

	// ===== КНОПКА БЭКАПА =====
	backupButton := widget.NewButton("Сделать резервную копию", func() {

		if targetFolder == "" {
			dialog.ShowInformation("Ошибка", "Выберите папку назначения", w)
			return
		}

		var err error

		switch backupMode {
		case "file":
			if fileURI == nil {
				dialog.ShowInformation("Ошибка", "Выберите файл для копирования", w)
				return
			}
			err = svc.RunManualBackup(fileURI.Path(), targetFolder)

		case "folder":
			if sourceFolder == "" {
				dialog.ShowInformation("Ошибка", "Выберите папку для копирования", w)
				return
			}
			err = svc.RunFolderBackup(sourceFolder, targetFolder)
		}

		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Готово", "Резервная копия успешно создана", w)
	})

	// ===== КОМПОНОВКА =====
	return container.NewVBox(
		widget.NewLabelWithStyle("Резервное копирование", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		modeSelector,
		widget.NewSeparator(),
		fileBlock,
		sourceBlock,
		widget.NewSeparator(),
		targetBlock,
		widget.NewSeparator(),
		backupButton,
	)
}
