package ui

import (
	"backup_master/internal/model"
	"backup_master/internal/service"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewBackup(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	manualTab := container.NewTabItem(
		"Ручной",
		newManualBackupTab(svc, w),
	)

	autoTab := container.NewTabItem(
		"Автоматический",
		newAutoBackupTab(svc, w),
	)

	return container.NewAppTabs(
		manualTab,
		autoTab,
	)
}

func newManualBackupTab(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	var (
		backupMode   = "file" // file | folder
		fileURI      fyne.URI
		sourceFolder string
	)

	// ===== ВЫБОР ФАЙЛА =====
	fileLabel := widget.NewLabel("Файл не выбран")
	fileButton := widget.NewButton("Выбрать файл", func() {
		dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if f == nil {
				return
			}
			fileURI = f.URI()
			fileLabel.SetText(fileURI.Path())
		}, w)
	})
	fileBlock := container.NewVBox(fileLabel, fileButton)

	// ===== ВЫБОР ПАПКИ =====
	sourceLabel := widget.NewLabel("Папка не выбрана")
	sourceButton := widget.NewButton("Выбрать папку", func() {
		dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if f == nil {
				return
			}
			sourceFolder = f.Path()
			sourceLabel.SetText(sourceFolder)
		}, w)
	})
	sourceBlock := container.NewVBox(sourceLabel, sourceButton)
	sourceBlock.Hide()

	// ===== РЕЖИМ =====
	modeSelector := widget.NewRadioGroup(
		[]string{"Файл", "Папка"},
		func(v string) {
			if v == "Файл" {
				backupMode = "file"
				fileBlock.Show()
				sourceBlock.Hide()
			} else {
				backupMode = "folder"
				fileBlock.Hide()
				sourceBlock.Show()
			}
		},
	)
	modeSelector.SetSelected("Файл")

	// ===== КНОПКА =====
	backupButton := widget.NewButton("Создать резервную копию", func() {

		dst := svc.Settings.BackupRootPath
		if dst == "" {
			dialog.ShowInformation(
				"Ошибка",
				"Не задана папка хранения бэкапов (настройки)",
				w,
			)
			return
		}

		var err error

		switch backupMode {
		case "file":
			if fileURI == nil {
				dialog.ShowInformation("Ошибка", "Выберите файл", w)
				return
			}
			err = svc.RunManualBackup(fileURI.Path(), dst)

		case "folder":
			if sourceFolder == "" {
				dialog.ShowInformation("Ошибка", "Выберите папку", w)
				return
			}
			err = svc.RunManualFolderBackup(sourceFolder, dst)
		}

		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation(
			"Готово",
			"Резервная копия успешно создана",
			w,
		)
	})

	return container.NewVBox(
		widget.NewLabelWithStyle(
			"Ручной бэкап",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		modeSelector,
		widget.NewSeparator(),
		fileBlock,
		sourceBlock,
		widget.NewSeparator(),
		backupButton,
	)
}

func newAutoBackupTab(svc *service.AppService, w fyne.Window) fyne.CanvasObject {

	var (
		tasks         []model.Task
		selectedIndex = -1
		list          *widget.List
	)

	loadTasks := func() {
		tasks, _ = svc.TaskRepo.GetAll()
		if list != nil {
			list.Refresh()
		}
	}

	list = widget.NewList(
		func() int {
			return len(tasks)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				Title("name"),
				widget.NewLabel("schedule"),
				widget.NewLabel("next"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)

			c.Objects[0].(*widget.Label).SetText(tasks[id].Name)
			c.Objects[1].(*widget.Label).SetText(
				HumanizeCron(tasks[id].Schedule),
			)
			c.Objects[2].(*widget.Label).SetText(
				"Следующий запуск: " + NextRunString(tasks[id]),
			)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}

	loadTasks()

	// ===== ПРОГРЕСС =====
	progressBar := widget.NewProgressBar()
	progressLabel := widget.NewLabel("")

	go func() {
		for p := range svc.Progress {
			fyne.Do(func() {
				progressBar.SetValue(float64(p.Percent) / 100)
				progressLabel.SetText(p.Message)
			})
		}
	}()

	// ===== КНОПКИ =====
	addBtn := widget.NewButton("Добавить правило", func() {
		showCreateTaskDialog(svc, w, loadTasks)
	})

	runBtn := widget.NewButton("Запустить сейчас", func() {
		if selectedIndex < 0 {
			return
		}
		go svc.RunTask(tasks[selectedIndex])
	})

	toggleBtn := widget.NewButton("Вкл / Выкл", func() {
		if selectedIndex < 0 {
			return
		}
		t := tasks[selectedIndex]
		_ = svc.TaskRepo.SetEnabled(t.ID, !t.Enabled)
		svc.Scheduler.Reload()
		loadTasks()
	})

	deleteBtn := widget.NewButton("Удалить", func() {
		if selectedIndex < 0 {
			return
		}
		_ = svc.TaskRepo.Delete(tasks[selectedIndex].ID)
		svc.Scheduler.Reload()
		selectedIndex = -1
		loadTasks()
	})

	return container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			progressLabel,
			progressBar,
		),
		nil,
		container.NewVBox(
			addBtn,
			runBtn,
			toggleBtn,
			deleteBtn,
		),
		list,
	)
}

func showCreateTaskDialog(
	svc *service.AppService,
	w fyne.Window,
	onSave func(),
) {
	// ===== ОБЩИЕ =====
	nameEntry := widget.NewEntry()

	// ===== ИСТОЧНИК =====
	var (
		sourcePath string
		sourceType = "file"
	)

	sourceLabel := widget.NewLabel("Не выбран")

	sourceTypeRadio := widget.NewRadioGroup(
		[]string{"Файл", "Папка"},
		func(v string) {
			if v == "Файл" {
				sourceType = "file"
			} else {
				sourceType = "folder"
			}
			sourcePath = ""
			sourceLabel.SetText("Не выбран")
		},
	)
	sourceTypeRadio.SetSelected("Файл")

	selectSourceBtn := widget.NewButton("Выбрать источник", func() {
		if sourceType == "file" {
			dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
				if err != nil || f == nil {
					return
				}
				sourcePath = f.URI().Path()
				sourceLabel.SetText(sourcePath)
			}, w)
		} else {
			dialog.ShowFolderOpen(func(f fyne.ListableURI, err error) {
				if err != nil || f == nil {
					return
				}
				sourcePath = f.Path()
				sourceLabel.SetText(sourcePath)
			}, w)
		}
	})

	// ===== РАСПИСАНИЕ =====
	periodSelect := widget.NewSelect(
		[]string{
			"Каждый час",
			"Каждый день",
			"Каждую неделю",
			"Каждый месяц",
		},
		nil,
	)

	minuteSelect := widget.NewSelect(genRange(0, 59), nil)
	hourSelect := widget.NewSelect(genRange(0, 23), nil)
	weekdaySelect := widget.NewSelect(
		[]string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"},
		nil,
	)
	dayOfMonthSelect := widget.NewSelect(genRange(1, 31), nil)

	scheduleBox := container.NewVBox()

	periodSelect.OnChanged = func(v string) {
		scheduleBox.Objects = nil

		switch v {
		case "Каждый час":
			scheduleBox.Add(
				widget.NewForm(
					widget.NewFormItem("Минута", minuteSelect),
				),
			)

		case "Каждый день":
			scheduleBox.Add(
				widget.NewForm(
					widget.NewFormItem("Час", hourSelect),
					widget.NewFormItem("Минута", minuteSelect),
				),
			)

		case "Каждую неделю":
			scheduleBox.Add(
				widget.NewForm(
					widget.NewFormItem("День недели", weekdaySelect),
					widget.NewFormItem("Час", hourSelect),
					widget.NewFormItem("Минута", minuteSelect),
				),
			)

		case "Каждый месяц":
			scheduleBox.Add(
				widget.NewForm(
					widget.NewFormItem("Дата", dayOfMonthSelect),
					widget.NewFormItem("Час", hourSelect),
					widget.NewFormItem("Минута", minuteSelect),
				),
			)
		}

		scheduleBox.Refresh()
	}

	periodSelect.SetSelected("Каждый день")

	// ===== КНОПКИ =====
	saveBtn := widget.NewButton("Создать", func() {

		if nameEntry.Text == "" || sourcePath == "" {
			dialog.ShowInformation("Ошибка", "Заполните все поля", w)
			return
		}

		cron, err := service.BuildCron(
			periodSelect.Selected,
			minuteSelect.Selected,
			hourSelect.Selected,
			weekdaySelect.Selected,
			dayOfMonthSelect.Selected,
		)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		task := &model.Task{
			Name:       nameEntry.Text,
			SourcePath: sourcePath,
			SourceType: sourceType,
			Schedule:   cron,
			Enabled:    true,
		}

		if err := svc.TaskRepo.Create(task); err != nil {
			dialog.ShowError(err, w)
			return
		}

		svc.Scheduler.Reload()

		onSave()
		dialog.ShowInformation("Готово", "Правило создано", w)
	})

	cancelBtn := widget.NewButton("Отмена", func() {
		w.Canvas().Overlays().Top().Hide()
	})

	// ===== ДИАЛОГ =====
	content := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Название", nameEntry),
		),
		widget.NewSeparator(),
		sourceTypeRadio,
		sourceLabel,
		selectSourceBtn,
		widget.NewSeparator(),
		widget.NewLabel("Расписание"),
		periodSelect,
		scheduleBox,
		widget.NewSeparator(),
		container.NewHBox(
			layout.NewSpacer(),
			cancelBtn,
			saveBtn,
		),
	)

	d := dialog.NewCustom(
		"Новое правило бэкапа",
		"Закрыть",
		content,
		w,
	)
	d.Resize(fyne.NewSize(420, 520))
	d.Show()
}

func genRange(from, to int) []string {
	var out []string
	for i := from; i <= to; i++ {
		out = append(out, fmt.Sprintf("%02d", i))
	}
	return out
}
