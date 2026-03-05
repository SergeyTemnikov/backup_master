package ui

import (
	"backup_master/internal/model"
	"backup_master/internal/service"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Dmitriy147/fynecalendar"
)

const (
	all           = "Все"
	successStatus = "OK"
	errorStatus   = "ERROR"
)

var listFilter string = all
var list *widget.List
var items []model.BackupWithTask
var selectedDate *time.Time

func NewBackupHistory(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	list = widget.NewList(
		func() int {
			return len(items)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				container.NewHBox(
					widget.NewLabel("Задача"),
					widget.NewSeparator(),
					widget.NewLabel("Статус"),
					widget.NewSeparator(),
					widget.NewLabel("Размер"),
				),
				container.NewHBox(
					widget.NewLabel("Время"),
					widget.NewLabel(""),
				),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			b := items[id]
			c := obj.(*fyne.Container)

			header := c.Objects[0].(*fyne.Container)
			header.Objects[0].(*widget.Label).SetText(b.TaskName)
			header.Objects[2].(*widget.Label).SetText(b.Status)
			header.Objects[4].(*widget.Label).SetText(formatBytes(b.SizeBytes))

			footer := c.Objects[1].(*fyne.Container)
			footer.Objects[0].(*widget.Label).SetText(
				b.StartedAt.Format("02.01.2006 15:04"),
			)

			errorLabel := footer.Objects[1].(*widget.Label)
			if b.ErrorMsg != nil && strings.TrimSpace(*b.ErrorMsg) != "" {
				errorLabel.SetText("Ошибка: " + *b.ErrorMsg)
			} else {
				errorLabel.SetText("")
			}
		},
	)
	update := func() {
		err := refresh(svc)
		if err != nil {
			dialog.ShowError(err, w)
		}
	}

	startListUpdate(w, 3*time.Second, update)

	return container.NewBorder(
		container.NewVBox(
			Title("История резервных копий"),
			container.NewHBox(
				NewFilterPopUP(svc, w),
				NewDateFilterButton(svc, w),
			),
		),
		nil,
		nil,
		nil,
		list,
	)
}

func refresh(svc *service.AppService) error {
	filter := listFilter
	if listFilter == all {
		filter = ""
	}

	var dateFilter *time.Time
	if selectedDate != nil {
		dateFilter = selectedDate
	}

	data, err := svc.GetBackupHistory(200, filter, dateFilter)
	if err != nil {
		return err
	}
	items = data
	list.Refresh()
	return nil
}

func NewFilterPopUP(svc *service.AppService, w fyne.Window) fyne.CanvasObject {
	filterSelect := widget.NewSelect(
		[]string{
			all,
			successStatus,
			errorStatus,
		},
		nil,
	)

	filterSelect.Selected = listFilter

	filterSelect.OnChanged = func(v string) {
		listFilter = v
		err := refresh(svc)
		if err != nil {
			dialog.ShowError(err, w)
		}
	}

	return container.NewHBox(widget.NewLabel("Статус: "), filterSelect)
}

type date struct {
	instruction *widget.Label
	dateChosen  *widget.Label
}

func NewDateFilterButton(svc *service.AppService, w fyne.Window) fyne.CanvasObject {
	dateBtn := widget.NewButton("Выбрать дату", nil)

	dateBtn.OnTapped = func() {

		var pop *widget.PopUp

		var initialDate time.Time
		if selectedDate != nil {
			initialDate = *selectedDate
		} else {
			initialDate = time.Now()
		}

		calendar := fynecalendar.NewMyCalendar(
			true,
			initialDate,
			time.Now().AddDate(-1, 0, 0),
			time.Now(),
			func(t time.Time) {
				selected := t
				selectedDate = &selected
				dateBtn.SetText(t.Format("02.01.2006"))
				err := refresh(svc)
				if err != nil {
					dialog.ShowError(err, w)
				}
				pop.Hide()
			},
		)

		content := container.NewVBox(
			widget.NewLabel("Выберите дату"),
			calendar,
			widget.NewButton("Сбросить", func() {
				selectedDate = nil
				dateBtn.SetText("Выбрать дату")
				err := refresh(svc)
				if err != nil {
					dialog.ShowError(err, w)
				}
				pop.Hide()
			}),
			widget.NewButton("Закрыть", func() {
				pop.Hide()
			}),
		)

		pop = widget.NewModalPopUp(content, w.Canvas())
		pop.Show()
	}

	return dateBtn
}
