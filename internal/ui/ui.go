package ui

import (
	"backup_master/internal/service"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func LoadUI(app fyne.App, svc *service.AppService) {
	w := app.NewWindow("Backup Master")
	w.Resize(fyne.NewSize(1280, 720))

	title := canvas.NewText("Backup Master", color.White)
	title.TextSize = 32
	title.TextStyle.Bold = true

	tabs := container.NewAppTabs(
		container.NewTabItem("Главная", NewDashboard(svc)),
		container.NewTabItem("Бэкап", NewBackup(svc, w)),
		container.NewTabItem("Восстановление", NewRestore(svc, w)),
		// container.NewTabItem("Планы", NewPlanner(svc)),
		// container.NewTabItem("Восстановление", NewRestore(svc)),
	)

	w.SetContent(container.NewBorder(title, nil, nil, nil, tabs))
	w.Show()
}
