package main

import (
	"backup_master/internal/service"
	"backup_master/internal/ui"
	"log"

	"fyne.io/fyne/v2/app"
)

func main() {
	svc, err := service.NewAppService("data/backup.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := svc.EnsureDemoData(); err != nil {
		log.Fatal(err)
	}

	a := app.New()
	ui.LoadUI(a, svc)
	a.Run()
}
