package ui

import (
	"backup_master/internal/model"
	"backup_master/internal/service"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewBackupHistory(svc *service.AppService, _ fyne.Window) fyne.CanvasObject {

	var items []model.BackupWithTask

	load := func() {
		items, _ = svc.GetBackupHistory(200)
	}

	load()

	list := widget.NewList(
		func() int {
			return len(items)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				container.NewHBox(
					widget.NewLabel("Задача"),
					widget.NewLabel("Статус"),
					widget.NewLabel("Размер"),
				),
				widget.NewLabel("Время"),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			b := items[id]
			c := obj.(*fyne.Container)

			header := c.Objects[0].(*fyne.Container)
			header.Objects[0].(*widget.Label).SetText(b.TaskName)
			header.Objects[1].(*widget.Label).SetText(b.Status)
			header.Objects[2].(*widget.Label).SetText(formatBytes(b.SizeBytes))

			c.Objects[1].(*widget.Label).SetText(
				b.StartedAt.Format("02.01.2006 15:04"),
			)

			errorLabel := c.Objects[2].(*widget.Label)
			if b.ErrorMsg != nil && strings.TrimSpace(*b.ErrorMsg) != "" {
				errorLabel.SetText("Ошибка: " + *b.ErrorMsg)
			} else {
				errorLabel.SetText("")
			}
		},
	)

	return container.NewBorder(
		Title("История резервных копий"),
		nil,
		nil,
		nil,
		list,
	)
}
