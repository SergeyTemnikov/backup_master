package ui

import (
	"backup_master/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var fileButton *widget.Button
var selectedFile *widget.Label
var fileURI fyne.URI

var folderButton *widget.Button
var selectedFolder *widget.Label
var folderPath string

func NewBackup(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	fileLabel := widget.NewLabel("Выберите файл для копирования: ")
	fileButton = widget.NewButton("Выбрать файл", func() { showFilePicker(w) })
	selectedFile = widget.NewLabel("Файл не выбран")

	fileCont := container.NewVBox(fileLabel, fileButton, selectedFile)

	folderLabel := widget.NewLabel("Выберите папку под копию: ")
	folderButton = widget.NewButton("Выбрать папку", func() { showFolderPicker(w) })
	selectedFolder = widget.NewLabel("Папка не выбрана")

	folderCont := container.NewVBox(folderLabel, folderButton, selectedFolder)

	backupButton := widget.NewButton("Сделать резервную копию", func() {
		if fileURI == nil || folderPath == "" {
			dialog.ShowInformation("Ошибка", "Выберите файл и папку", w)
			return
		}

		err := svc.RunManualBackup(fileURI.Path(), folderPath)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Готово", "Резервная копия успешно создана", w)
	})

	return container.NewVBox(fileCont, folderCont, backupButton)
}

func showFilePicker(w fyne.Window) {
	dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
		saveFile := "NoFileYet"
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if f == nil {
			return
		}
		saveFile = f.URI().Path()
		fileURI = f.URI()
		selectedFile.SetText(saveFile)
	}, w)
}

func showFolderPicker(w fyne.Window) {
	dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
		saveFolder := "NoFolderYet"
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if f == nil {
			return
		}
		saveFolder = f.Path()
		folderPath = saveFolder
		selectedFolder.SetText(saveFolder)
	}, w)
}
