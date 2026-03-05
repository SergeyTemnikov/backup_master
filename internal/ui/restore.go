package ui

import (
	"backup_master/internal/model"
	"backup_master/internal/service"
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var backupList *widget.List
var backups []model.BackupWithTask

func NewRestore(svc *service.AppService, w fyne.Window) fyne.CanvasObject {
	var selected *model.BackupWithTask
	var targetDir string

	taskLabel := widget.NewLabel("-")
	statusLabel := widget.NewLabel("-")
	sizeLabel := widget.NewLabel("-")
	dateLabel := widget.NewLabel("-")
	backupPathLabel := widget.NewLabel("-")

	targetLabel := widget.NewLabel("Не выбрана")

	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	restoreBtn := widget.NewButton("Восстановить", nil)
	restoreBtn.Disable()

	selectDirBtn := widget.NewButton("Выберите папку назначения", func() {

		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {

			if uri == nil {
				return
			}

			targetDir = uri.Path()
			targetLabel.SetText(targetDir)

			if selected != nil {
				restoreBtn.Enable()
			}

		}, w)

	})

	overwriteCheck := widget.NewCheck(
		"Перезаписать существующие файлы",
		func(v bool) {
			if selected == nil {
				return
			}

			if v {
				targetDir = filepath.Dir(selected.SourcePath)
				targetLabel.SetText(targetDir)
				restoreBtn.Enable()
				selectDirBtn.Disable()

			} else {
				targetDir = ""
				targetLabel.SetText("Не выбрана")
				restoreBtn.Disable()
				selectDirBtn.Enable()
			}
		},
	)

	backupList = widget.NewList(

		func() int {
			return len(backups)
		},

		func() fyne.CanvasObject {
			return widget.NewLabel("backup")
		},

		func(i widget.ListItemID, o fyne.CanvasObject) {

			b := backups[i]

			text := fmt.Sprintf(
				"%d | %s | %s",
				b.ID,
				b.TaskName,
				b.Status,
			)

			o.(*widget.Label).SetText(text)
		},
	)

	backupList.OnSelected = func(id widget.ListItemID) {

		b := backups[id]
		selected = &b

		taskLabel.SetText(b.TaskName)
		statusLabel.SetText(b.Status)
		sizeLabel.SetText(formatSize(b.SizeBytes))
		dateLabel.SetText(b.FinishedAt.Format("2006-01-02 15:04"))
		backupPathLabel.SetText(b.TargetPath)

		if overwriteCheck.Checked {

			targetDir = filepath.Dir(b.SourcePath)
			targetLabel.SetText(targetDir)

			restoreBtn.Enable()

		} else if targetDir != "" {

			restoreBtn.Enable()
		}
	}

	restoreBtn.OnTapped = func() {

		if selected == nil {
			return
		}

		if targetDir == "" {
			dialog.ShowInformation("Выберите папку", "Сначала выберите папка назначение", w)
			return
		}

		dialog.ShowConfirm(

			"Подтвердить восстановление",

			fmt.Sprintf(
				"Восстановление #%d?\nЗадача: %s",
				selected.ID,
				selected.TaskName,
			),

			func(ok bool) {

				if !ok {
					return
				}

				progress.Show()

				go func() {

					var err error

					backupPath := selected.TargetPath

					if selected.SourceType == "file" {

						err = svc.RunFileRestoreWithChecksum(
							selected.ID,
							targetDir,
							overwriteCheck.Checked,
						)

					} else {

						err = svc.RunFolderRestore(
							filepath.Dir(backupPath),
							targetDir,
							overwriteCheck.Checked,
						)

					}

					fyne.Do(progress.Hide)

					if err != nil {

						dialog.ShowError(err, w)
						return
					}

					dialog.ShowInformation(
						"Восстановление выполнено",
						"Резервная копия успешно восстановлена",
						w,
					)

				}()

			},

			w,
		)

	}

	info := container.NewVBox(

		widget.NewLabelWithStyle("Информация о копии", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		formRow("Задача", taskLabel),
		formRow("Статус", statusLabel),
		formRow("Размер", sizeLabel),
		formRow("Закончен", dateLabel),
		formRow("Путь", backupPathLabel),

		widget.NewSeparator(),

		widget.NewLabel("Папка назначения"),
		targetLabel,

		selectDirBtn,
		overwriteCheck,

		widget.NewSeparator(),

		restoreBtn,
		progress,
	)

	left := container.NewBorder(
		widget.NewLabel("Копии"),
		nil,
		nil,
		nil,
		backupList,
	)

	split := container.NewHSplit(left, info)
	split.SetOffset(0.45)

	update := func() {
		err := refreshList(svc)
		if err != nil {
			dialog.ShowError(err, w)
		}
	}

	startListUpdate(w, 3*time.Second, update)

	return split
}

func refreshList(svc *service.AppService) error {
	data, err := svc.GetBackupHistory(200, "OK", nil)
	if err != nil {
		return err
	}

	backups = data
	backupList.Refresh()
	return nil
}

func formRow(name string, value fyne.CanvasObject) fyne.CanvasObject {

	return container.NewBorder(
		nil,
		nil,
		widget.NewLabel(name),
		nil,
		value,
	)
}

func formatSize(bytes int64) string {

	const unit = 1024

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0

	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB",
		float64(bytes)/float64(div),
		"KMGTPE"[exp],
	)
}

func startListUpdate(
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
