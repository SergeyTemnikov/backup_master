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

	success := widget.NewLabel("0")
	errors := widget.NewLabel("0")
	upcoming := widget.NewLabel("0")

	update := func() {
		success.SetText(strconv.Itoa(svc.SuccessCount()))
		errors.SetText(strconv.Itoa(svc.ErrorCount()))
		upcoming.SetText(strconv.Itoa(svc.UpcomingCount()))
	}

	update()

	startDashboardMonitor(w, 3*time.Second, update)

	status := container.NewHBox(
		container.NewVBox(Title("Успешно"), success),
		container.NewVBox(Title("Ошибки"), errors),
		container.NewVBox(Title("Ближайшие"), upcoming),
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

	return container.NewVScroll(
		container.NewVBox(
			Title("Статус"),
			status,
			container.NewVBox(
				Title("Хранилище"),
				bar,
				label,
			),
		),
	)
}

func startDashboardMonitor(
	w fyne.Window,
	interval time.Duration,
	update func(),
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			fyne.Do(update)
		}
	}()
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
