package ui

import (
	"backup_master/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

func LoadUI(app fyne.App, svc *service.AppService) {
	w := app.NewWindow("Backup Master")
	w.Resize(fyne.NewSize(900, 600))

	setupTray(app, w)

	w.SetCloseIntercept(func() {
		w.Hide()
	})

	settingsTab := container.NewTabItem("Настройки", NewSettings(svc, w))

	tabs := container.NewAppTabs(
		container.NewTabItem("Главная", NewDashboard(svc, w)),
		container.NewTabItem("Бэкап", NewBackup(svc, w)),
		container.NewTabItem("Восстановление", NewRestore(svc, w)),
		container.NewTabItem("История", NewBackupHistory(svc, w)),
		settingsTab,
	)

	w.SetContent(container.NewBorder(Title("Backup Master"), nil, nil, nil, tabs))
	w.Show()

	if svc.Settings.BackupRootPath == "" {
		ShowFirstRunDialog(svc, w, func() {
			settingsTab.Content = NewSettings(svc, w)
			settingsTab.Content.Refresh()
		})
	}
}

func setupTray(app fyne.App, w fyne.Window) {
	if desk, ok := app.(desktop.App); ok {

		showItem := fyne.NewMenuItem("Открыть", func() {
			w.Show()
		})

		hideItem := fyne.NewMenuItem("Скрыть", func() {
			w.Hide()
		})

		quitItem := fyne.NewMenuItem("Выход", func() {
			app.Quit()
		})

		menu := fyne.NewMenu(
			"Backup Master",
			showItem,
			hideItem,
			fyne.NewMenuItemSeparator(),
			quitItem,
		)

		desk.SetSystemTrayMenu(menu)

		desk.SetSystemTrayIcon(theme.ComputerIcon())
	}
}
