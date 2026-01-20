package ui

import (
	"backup_master/internal/service"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewDashboard(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	status := container.NewHBox(
		statusCard("Успешно", svc.SuccessCount()),
		statusCard("Ошибки", svc.ErrorCount()),
		statusCard("Ближайшие", svc.UpcomingCount()),
	)

	bar := progressBar(svc)
	label := widget.NewLabel("")

	startStorageMonitor(
		bar,
		label,
		w,
		func() (int64, error) {
			return svc.GetStorageUsedBytes()
		},
		svc.Settings.MaxStorageBytes,
	)

	storageBlock := container.NewVBox(
		widget.NewLabel("Хранилище"),
		bar,
		label,
	)

	return container.NewVScroll(
		container.NewVBox(
			Title("Статус"),
			status,
			layout.NewSpacer(),
			storageBlock,
		),
	)
}

func progressBar(svc *service.AppService) *widget.ProgressBar {
	bar := widget.NewProgressBar()
	label := widget.NewLabel("")

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			used, _ := svc.StorageRepo.CalcDirSize(svc.Settings.BackupRootPath)
			total := svc.Settings.MaxStorageBytes

			bar.Max = float64(total)
			bar.SetValue(float64(used))
			label.SetText(formatBytes(used) + " / " + formatBytes(total))
		}
	}()

	return bar
}

func statusCard(title string, count int) fyne.CanvasObject {
	return container.NewVBox(
		Title(title),
		Title(strconv.Itoa(count)),
	)
}

// TODO: Логика продолжения бэкапа и стартовое окно
func startStorageMonitor(
	bar *widget.ProgressBar,
	label *widget.Label,
	w fyne.Window,
	getUsed func() (int64, error),
	maxBytes int64,
) {
	var limitDialogShown bool

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			used, err := getUsed()
			if err != nil {
				continue
			}

			bar.Max = float64(maxBytes)
			bar.SetValue(float64(used))
			label.SetText(
				formatBytes(used) + " / " + formatBytes(maxBytes),
			)

			// 🔴 Проверка превышения лимита
			if used > maxBytes && !limitDialogShown {
				limitDialogShown = true

				// Диалог должен создаваться в UI-контексте
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Превышен лимит хранилища",
					Content: "Занято больше места, чем разрешено в настройках",
				})

				dialog.ShowConfirm(
					"Превышен лимит",
					"Лимит хранилища превышен. Продолжить резервное копирование?",
					func(ok bool) {
						if ok {
							// продолжить
						}
					},
					w,
				)

			}

			if used <= maxBytes {
				limitDialogShown = false
			}
		}
	}()
}
