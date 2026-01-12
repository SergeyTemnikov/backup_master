package ui

import (
	"backup_master/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	restoreMode         = "file" // file | folder
	restoreFileURI      fyne.URI
	restoreSourceFolder string
	restoreTargetFolder string
	overwriteOriginal   bool
)

func NewRestore(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	// ===== ВЫБОР ФАЙЛА =====
	fileLabel := widget.NewLabel("Выберите файл резервной копии:")
	fileButton := widget.NewButton("Выбрать файл", func() {
		dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil || f == nil {
				return
			}
			restoreFileURI = f.URI()
			fileLabel.SetText("Файл: " + restoreFileURI.Path())
		}, w)
	})

	fileBlock := container.NewVBox(fileLabel, fileButton)

	// ===== ВЫБОР ПАПКИ ИСТОЧНИКА =====
	folderLabel := widget.NewLabel("Выберите папку резервной копии:")
	folderButton := widget.NewButton("Выбрать папку", func() {
		dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
			if err != nil || f == nil {
				return
			}
			restoreSourceFolder = f.Path()
			folderLabel.SetText("Папка: " + restoreSourceFolder)
		}, w)
	})

	folderBlock := container.NewVBox(folderLabel, folderButton)
	folderBlock.Hide()

	// ===== РЕЖИМ ВОССТАНОВЛЕНИЯ =====
	modeSelector := widget.NewRadioGroup(
		[]string{"Файл", "Папка"},
		func(s string) {
			if s == "Файл" {
				restoreMode = "file"

				fileBlock.Show()
				folderBlock.Hide()
			} else {
				restoreMode = "folder"

				fileBlock.Hide()
				folderBlock.Show()
			}
		},
	)
	modeSelector.SetSelected("Файл")

	// ===== ВЫБОР ПАПКИ НАЗНАЧЕНИЯ =====
	targetLabel := widget.NewLabel("Папка для восстановления:")
	targetBtn := widget.NewButton("Выбрать папку", func() {
		dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
			if err != nil || f == nil {
				return
			}
			restoreTargetFolder = f.Path()
			targetLabel.SetText("Папка: " + restoreTargetFolder)
		}, w)
	})

	targetBlock := container.NewVBox(targetLabel, targetBtn)

	// ===== ПЕРЕЗАПИСЬ =====
	overwriteCheck := widget.NewCheck("Перезаписать в исходную папку", func(v bool) {
		overwriteOriginal = v
		targetBlock.Hidden = v
		targetBlock.Refresh()
	})

	// ===== КНОПКА RESTORE =====
	restoreBtn := widget.NewButton("Восстановить", func() {

		if !overwriteOriginal && restoreTargetFolder == "" {
			dialog.ShowInformation("Ошибка", "Выберите папку назначения", w)
			return
		}

		var err error

		if restoreMode == "file" {
			if restoreFileURI == nil {
				dialog.ShowInformation("Ошибка", "Выберите файл", w)
				return
			}
			err = svc.RunFileRestore(
				restoreFileURI.Path(),
				restoreTargetFolder,
				overwriteOriginal,
			)
		} else {
			if restoreSourceFolder == "" {
				dialog.ShowInformation("Ошибка", "Выберите папку", w)
				return
			}
			err = svc.RunFolderRestore(
				restoreSourceFolder,
				restoreTargetFolder,
				overwriteOriginal,
			)
		}

		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Готово", "Восстановление успешно завершено", w)
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("Восстановление", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		modeSelector,
		widget.NewSeparator(),
		fileBlock,
		folderBlock,
		widget.NewSeparator(),
		overwriteCheck,
		targetBlock,
		widget.NewSeparator(),
		restoreBtn,
	)
}
