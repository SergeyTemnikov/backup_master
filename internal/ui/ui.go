package ui

import (
	"backup_master/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func LoadUI(app fyne.App, svc *service.AppService) {
	w := app.NewWindow("Backup Master")
	w.Resize(fyne.NewSize(800, 600))

	tabs := container.NewAppTabs(
		container.NewTabItem("Главная", NewDashboard(svc, w)),
		container.NewTabItem("Бэкап", NewBackup(svc, w)),
		container.NewTabItem("Восстановление", NewRestore(svc, w)),
		container.NewTabItem("Настройки", NewSettings(svc, w)),
	)

	w.SetContent(container.NewBorder(Title("Backup Master"), nil, nil, nil, tabs))
	w.Show()
}
