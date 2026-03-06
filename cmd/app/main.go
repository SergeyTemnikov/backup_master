package main

import (
	"backup_master/internal/service"
	"backup_master/internal/ui"
	"log"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	// Инициализация сервиса
	svc, err := service.NewAppService(a, "data/backup.db")
	if err != nil {
		log.Fatal(err)
	}

	// Запуск планировщика автобэкапов
	if err := svc.StartScheduler(); err != nil {
		log.Fatal(err)
	}

	// UI
	ui.LoadUI(a, svc)
	a.Run()
}
