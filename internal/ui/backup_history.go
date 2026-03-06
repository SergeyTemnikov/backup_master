package ui

import (
	"backup_master/internal/model"
	"backup_master/internal/service"
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
var selectedTaskID *int64
var RED = color.RGBA{R: 255, G: 0, B: 0, A: 100}
var GREEN = color.RGBA{R: 0, G: 255, B: 0, A: 100}

func NewBackupHistory(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	list = widget.NewList(
		func() int {
			return len(items)
		},
		func() fyne.CanvasObject {
			errorText := widget.NewRichTextFromMarkdown("")
			errorText.Wrapping = fyne.TextWrapWord

			return container.NewGridWithColumns(
				2,
				container.NewVBox(
					container.NewHBox(
						widget.NewLabel("Задача"),
						widget.NewSeparator(),
						canvas.NewText("Статус", RED),
						widget.NewSeparator(),
						widget.NewLabel("Размер"),
					),
					widget.NewLabel("Время"),
				),
				container.NewScroll(errorText),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			b := items[id]
			c := obj.(*fyne.Container)

			left := c.Objects[0].(*fyne.Container)
			header := left.Objects[0].(*fyne.Container)
			header.Objects[0].(*widget.Label).SetText(b.TaskName)

			status := header.Objects[2].(*canvas.Text)
			status.Text = b.Status
			status.TextStyle = fyne.TextStyle{Bold: true}

			if b.Status == "OK" {
				status.Color = GREEN
			} else {
				status.Color = RED
			}

			header.Objects[4].(*widget.Label).SetText(formatBytes(b.SizeBytes))

			left.Objects[1].(*widget.Label).SetText(
				b.StartedAt.Format("02.01.2006 15:04"),
			)

			errorLabel := c.Objects[1].(*container.Scroll).Content.(*widget.RichText)
			if b.ErrorMsg != nil && strings.TrimSpace(*b.ErrorMsg) != "" {
				errorLabel.ParseMarkdown("**Ошибка:** " + *b.ErrorMsg)
			} else {
				errorLabel.ParseMarkdown("")
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

	exportBtn := widget.NewButton("Экспорт", func() {

		save := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {

			if writer == nil {
				return
			}

			logs, err := svc.GetBackupHistory(1000, "", nil, nil)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			for _, l := range logs {

				line := fmt.Sprintf(
					"%s | %s | %s | %s\n",
					l.StartedAt.Format("2006-01-02 15:04"),
					l.TaskName,
					l.Status,
					l.TargetPath,
				)

				writer.Write([]byte(line))
			}

			writer.Close()

		}, w)

		save.SetFileName("backup_logs.txt")

		save.Show()

	})

	return container.NewBorder(
		container.NewVBox(
			Title("История резервных копий"),
			container.NewHBox(
				NewFilterPopUP(svc, w),
				NewDateFilterButton(svc, w),
				NewTaskFilter(svc, w),
				exportBtn,
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

	data, err := svc.GetBackupHistory(200, filter, dateFilter, selectedTaskID)
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

func NewTaskFilter(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	tasks, err := svc.TaskRepo.GetAll()
	if err != nil {
		return widget.NewLabel("Ошибка загрузки задач")
	}

	var manualBackupId int64 = -1
	optionsMap := map[string]*int64{"Все задачи": nil, "Ручной бэкап": &manualBackupId}
	options := []string{"Все задачи", "Ручной бэкап"}

	for _, t := range tasks {
		options = append(options, t.Name)
		optionsMap[t.Name] = &t.ID
	}

	selectTask := widget.NewSelect(options, nil)

	selectTask.OnChanged = func(v string) {

		if v == "Все задачи" {
			selectedTaskID = nil
		} else {
			id := optionsMap[v]
			selectedTaskID = id
		}

		err := refresh(svc)
		if err != nil {
			dialog.ShowError(err, w)
		}
	}

	selectTask.Selected = "Все задачи"

	return container.NewHBox(
		widget.NewLabel("Задача: "),
		selectTask,
	)
}
