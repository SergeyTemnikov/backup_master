package ui

import (
	"backup_master/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func LoadUI(app fyne.App, svc *service.AppService) {
	w := app.NewWindow("Backup Master")
	w.Resize(fyne.NewSize(1280, 720))

	topBar := container.NewHBox(
		Title("Backup Master"),
		layout.NewSpacer(),
		ThemeToggleButton(),
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("Главная", NewDashboard(svc)),
		container.NewTabItem("Бэкап", NewBackup(svc, w)),
		container.NewTabItem("Восстановление", NewRestore(svc, w)),
		// container.NewTabItem("Планы", NewPlanner(svc)),
		// container.NewTabItem("Восстановление", NewRestore(svc)),
	)

	w.SetContent(container.NewBorder(topBar, nil, nil, nil, tabs))
	w.Show()
}
