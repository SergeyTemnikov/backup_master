package ui

import (
	"backup_master/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var restoreBackupPath string
var restoreTargetPath string

func NewRestore(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	backupLabel := widget.NewLabel("Выберите файл резервной копии:")
	backupBtn := widget.NewButton("Выбрать backup", func() {
		dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil || f == nil {
				return
			}
			restoreBackupPath = f.URI().Path()
			backupLabel.SetText("Backup: " + restoreBackupPath)
		}, w)
	})

	targetLabel := widget.NewLabel("Выберите папку для восстановления:")
	targetBtn := widget.NewButton("Выбрать папку", func() {
		dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
			if err != nil || f == nil {
				return
			}
			restoreTargetPath = f.Path()
			targetLabel.SetText("Папка: " + restoreTargetPath)
		}, w)
	})

	restoreBtn := widget.NewButton("Восстановить", func() {
		if restoreBackupPath == "" || restoreTargetPath == "" {
			dialog.ShowInformation("Ошибка", "Выберите backup и папку", w)
			return
		}

		if err := svc.RestoreBackup(restoreBackupPath, restoreTargetPath); err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Готово", "Файл успешно восстановлен", w)
	})

	return container.NewVBox(
		backupLabel,
		backupBtn,
		widget.NewSeparator(),
		targetLabel,
		targetBtn,
		widget.NewSeparator(),
		restoreBtn,
	)
}
