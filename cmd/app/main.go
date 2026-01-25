package main

import (
	"backup_master/internal/service"
	"backup_master/internal/ui"
	"log"

	"fyne.io/fyne/v2/app"
)

func main() {
	// Инициализация сервиса
	svc, err := service.NewAppService("data/backup.db")
	if err != nil {
		log.Fatal(err)
	}

	// Демоданные (один раз)
	// if err := svc.EnsureDemoData(); err != nil {
	// 	log.Fatal(err)
	// }

	// Запуск планировщика автобэкапов
	if err := svc.StartScheduler(); err != nil {
		log.Fatal(err)
	}

	// UI
	a := app.New()
	ui.LoadUI(a, svc)
	a.Run()
}
