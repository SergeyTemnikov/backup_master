package ui

import (
	"backup_master/internal/service"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewDashboard(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	status := container.NewHBox(
		statusCard("Успешно", svc.SuccessCount()),
		statusCard("Ошибки", svc.ErrorCount()),
		statusCard("Ближайшие", svc.UpcomingCount()),
	)

	bar := progressBar()
	label := widget.NewLabel("")

	startStorageMonitor(
		bar,
		label,
		w,
		func() (int64, error) {
			return svc.GetStorageUsedBytes()
		},
		func() int64 {
			return svc.Settings.MaxStorageBytes
		},
	)

	storageBlock := container.NewVBox(
		Title("Хранилище"),
		bar,
		label,
	)

	return container.NewVScroll(
		container.NewVBox(
			Title("Статус"),
			status,
			storageBlock,
		),
	)
}

func progressBar() *widget.ProgressBar {
	bar := widget.NewProgressBar()
	bar.Min = 0
	return bar
}

func statusCard(title string, count int) fyne.CanvasObject {
	return container.NewVBox(
		Title(title),
		Title(strconv.Itoa(count)),
	)
}

func startStorageMonitor(
	bar *widget.ProgressBar,
	label *widget.Label,
	w fyne.Window,
	getUsed func() (int64, error),
	getMax func() int64,
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

			maxBytes := getMax()

			fyne.Do(func() {
				bar.Max = float64(maxBytes)
				bar.SetValue(float64(used))
				label.SetText(
					formatBytes(used) + " / " + formatBytes(maxBytes),
				)

				if maxBytes > 0 && used > maxBytes && !limitDialogShown {
					limitDialogShown = true

					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Превышен лимит хранилища",
						Content: "Занято больше места, чем разрешено в настройках",
					})

					dialog.ShowConfirm(
						"Превышен лимит",
						"Лимит хранилища превышен. Продолжить резервное копирование?",
						func(ok bool) {},
						w,
					)
				}

				if used <= maxBytes {
					limitDialogShown = false
				}
			})
		}
	}()
}
