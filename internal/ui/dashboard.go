package ui

import (
	"backup_master/internal/service"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewDashboard(svc *service.AppService) fyne.CanvasObject {

	status := container.NewHBox(
		statusCard("Успешно", svc.SuccessCount()),
		statusCard("Ошибки", svc.ErrorCount()),
		statusCard("Ближайшие", svc.UpcomingCount()),
	)

	storages := container.NewVBox()
	for _, s := range svc.Storages() {
		bar := widget.NewProgressBar()
		bar.Max = float64(s.Total)
		bar.SetValue(float64(s.Used))
		storages.Add(container.NewVBox(
			widget.NewLabel(s.Name),
			bar,
		))
	}

	table := widget.NewTable(
		func() (int, int) { return len(svc.LastBackups()), 3 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			b := svc.LastBackups()[id.Row]
			l := o.(*widget.Label)
			switch id.Col {
			case 0:
				l.SetText("Task #" + strconv.FormatInt(b.TaskID, 10))
			case 1:
				l.SetText(b.StartedAt.Format("02.01 15:04"))
			case 2:
				l.SetText(b.Status)
			}
		},
	)
	//table.SetMinSize(fyne.NewSize(0, 200))

	return container.NewVScroll(container.NewVBox(
		Title("Статус"),
		status,
		layout.NewSpacer(),
		Title("Хранилища"),
		storages,
		layout.NewSpacer(),
		Title("Последние копии"),
		table,
	))
}

func statusCard(title string, count int) fyne.CanvasObject {
	return container.NewVBox(
		Title(title),
		Title(strconv.Itoa(count)),
	)
}
