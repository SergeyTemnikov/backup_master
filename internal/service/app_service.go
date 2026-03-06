package service

import (
	"backup_master/internal/model"
	"backup_master/internal/repository"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

type AppService struct {
	App          fyne.App
	DB           *sql.DB
	TaskRepo     *repository.TaskRepository
	BackupRepo   *repository.BackupRepository
	SettingsRepo *repository.SettingsRepository

	Settings *model.AppSettings

	BackupSvc  *BackupService
	RestoreSvc *RestoreService

	Progress chan *model.BackupProgress

	Scheduler *Scheduler

	TaskStates    map[int64]model.TaskState
	TaskStatesMu  sync.RWMutex
	TaskStateChan chan model.TaskStateEvent
}

func NewAppService(app fyne.App, dbPath string) (*AppService, error) {
	db, err := repository.OpenDB(dbPath)
	if err != nil {
		return nil, err
	}

	if err := repository.InitSchema(db); err != nil {
		return nil, err
	}

	settingsRepo := repository.NewSettingsRepository(db)

	settings, err := settingsRepo.Get()
	if err != nil {
		return nil, err
	}

	if settings.BackupRootPath != "" {
		if err := CleanupTempArtifacts(settings.BackupRootPath); err != nil {
			log.Printf("cleanup temp failed: %v\n", err)
		}
	}

	svc := &AppService{
		App:          app,
		DB:           db,
		TaskRepo:     repository.NewTaskRepository(db),
		BackupRepo:   repository.NewBackupRepository(db),
		SettingsRepo: settingsRepo,
		Settings:     settings,

		BackupSvc:  NewBackupService(),
		RestoreSvc: NewRestoreService(),

		Progress: make(chan *model.BackupProgress, 16),

		TaskStates:    make(map[int64]model.TaskState),
		TaskStateChan: make(chan model.TaskStateEvent, 16),
	}

	svc.Scheduler = NewScheduler(svc)

	return svc, nil
}

//////////////////////
// Уведмоления
//////////////////////

func (s *AppService) StartNotificationListener() {
	go func() {
		for event := range s.TaskStateChan {

			switch event.State {

			case model.TaskSuccess:
				s.App.SendNotification(&fyne.Notification{
					Title:   "Backup Master",
					Content: "Задача успешно завершена",
				})

			case model.TaskError:
				s.App.SendNotification(&fyne.Notification{
					Title:   "Backup Master",
					Content: "Ошибка выполнения задачи",
				})
			}
		}
	}()
}
func (s *AppService) NotifyError(err error) {
	s.App.SendNotification(&fyne.Notification{
		Title:   "Ошибка Backup Master",
		Content: err.Error(),
	})
}

//////////////////////
// Ручной бэкап
//////////////////////

func (s *AppService) RunManualBackup(srcFile, dstFolder string) error {
	started := time.Now()

	size, checksum, path, err := s.BackupSvc.BackupFile(srcFile, dstFolder)

	finished := time.Now()

	return s.saveManualBackup(
		nil,
		size,
		checksum,
		srcFile,
		"file",
		path,
		started,
		finished,
		err,
	)
}

func (s *AppService) RunManualFolderBackup(srcFile, dstFolder string) error {
	started := time.Now()

	size, checksum, path, err := s.BackupSvc.BackupFolder(srcFile, dstFolder)

	finished := time.Now()

	return s.saveManualBackup(
		nil,
		size,
		checksum,
		srcFile,
		"folder",
		path,
		started,
		finished,
		err,
	)
}

func (s *AppService) saveManualBackup(
	taskID *int64,
	size int64,
	checksum string,
	sourcePath string,
	sourceType string,
	targetPath string,
	started time.Time,
	finished time.Time,
	err error,
) error {

	status := "OK"
	var errMsg *string

	if err != nil {
		status = "ERROR"
		msg := err.Error()
		errMsg = &msg
	}

	backup := &model.Backup{
		TaskID:       taskID,
		Status:       status,
		SizeBytes:    size,
		SourcePath:   sourcePath,
		SourceType:   sourceType,
		TargetPath:   targetPath,
		ErrorMessage: errMsg,
		Checksum:     checksum,
		StartedAt:    started,
		FinishedAt:   finished,
	}

	if dbErr := s.BackupRepo.Create(backup); dbErr != nil {
		return dbErr
	}

	return err
}

//////////////////////
// Запуск планировщика
//////////////////////

func (s *AppService) StartScheduler() error {
	if s.Scheduler == nil {
		s.Scheduler = NewScheduler(s)
	}
	return s.Scheduler.Start()
}

func (s *AppService) runTask(task model.Task) {
	if !s.trySetRunning(task.ID) {
		log.Printf("task %d already running, skip\n", task.ID)
		return
	}

	s.notifyTaskState(task.ID, model.TaskRunning)

	finalState := model.TaskSuccess
	started := time.Now()

	defer func() {
		if r := recover(); r != nil {
			finalState = model.TaskError
		}
		s.setTaskState(task.ID, finalState)
		s.notifyTaskState(task.ID, finalState)
	}()

	select {
	case s.Progress <- &model.BackupProgress{
		TaskID:  task.ID,
		Percent: 0,
		Message: "Запуск задачи"}:
	default:
	}

	if err := s.CheckStorageLimit(); err != nil {
		finalState = model.TaskError
		s.sendTaskError(task.ID, err)
		return
	}

	var (
		size     int64
		err      error
		checksum string
		path     string
	)

	switch task.SourceType {
	case "file":
		size, checksum, path, err = s.BackupSvc.BackupFile(
			task.SourcePath,
			s.Settings.BackupRootPath,
		)

	case "folder":
		size, checksum, path, err = s.BackupSvc.BackupFolder(
			task.SourcePath,
			s.Settings.BackupRootPath,
		)

	default:
		finalState = model.TaskError
		s.sendTaskError(task.ID, fmt.Errorf("неизвестный тип источника"))
		return
	}

	status := "OK"
	var errMsg *string

	if err != nil {
		finalState = model.TaskError
		status = "ERROR"
		msg := err.Error()
		errMsg = &msg
	}

	backup := &model.Backup{
		TaskID:       &task.ID,
		Status:       status,
		SizeBytes:    size,
		SourcePath:   task.SourcePath,
		SourceType:   task.SourceType,
		TargetPath:   path,
		ErrorMessage: errMsg,
		Checksum:     checksum,
		StartedAt:    started,
		FinishedAt:   time.Now(),
	}

	if dbErr := s.BackupRepo.Create(backup); dbErr != nil {
		finalState = model.TaskError
		s.sendTaskError(task.ID, dbErr)
		return
	}
	select {
	case s.Progress <- &model.BackupProgress{
		TaskID:  task.ID,
		Percent: 100,
		Message: "Задача завершена"}:
	default:
	}
}

func (s *AppService) sendTaskError(taskID int64, err error) {
	s.NotifyError(err)

	select {
	case s.Progress <- &model.BackupProgress{
		TaskID:  taskID,
		Message: err.Error()}:
	default:
	}
}

func (s *AppService) CheckStorageLimit() error {
	used, err := s.GetUsedBytes(s.Settings.BackupRootPath)
	if err != nil {
		return err
	}

	if used > s.Settings.MaxStorageBytes {
		return fmt.Errorf("превышен лимит хранилища")
	}
	return nil
}

// ////////////////////
// Авто бэкап/таски
// ////////////////////

func (s *AppService) RunTask(task model.Task) {
	go s.runTask(task)
}

func (s *AppService) CreateTask(task *model.Task) error {
	err := s.TaskRepo.Create(task)
	if err == nil && s.Scheduler != nil {
		s.Scheduler.Reload()
	}
	return err
}

func (s *AppService) DeleteTask(taskID int64) error {
	err := s.TaskRepo.Delete(taskID)
	if err == nil && s.Scheduler != nil {
		s.Scheduler.Reload()
	}
	return err
}

func (s *AppService) SetTaskEnabled(taskID int64, enabled bool) error {
	err := s.TaskRepo.SetEnabled(taskID, enabled)
	if err == nil && s.Scheduler != nil {
		s.Scheduler.Reload()
	}
	return err
}

func (s *AppService) setTaskState(taskID int64, state model.TaskState) {
	s.TaskStatesMu.Lock()
	defer s.TaskStatesMu.Unlock()
	s.TaskStates[taskID] = state
}

func (s *AppService) trySetRunning(taskID int64) bool {
	s.TaskStatesMu.Lock()
	defer s.TaskStatesMu.Unlock()

	if st, ok := s.TaskStates[taskID]; ok && st == model.TaskRunning {
		return false
	}

	s.TaskStates[taskID] = model.TaskRunning
	return true
}

func (s *AppService) getTaskState(taskID int64) model.TaskState {
	s.TaskStatesMu.RLock()
	defer s.TaskStatesMu.RUnlock()

	if st, ok := s.TaskStates[taskID]; ok {
		return st
	}
	return model.TaskIdle
}

func (s *AppService) notifyTaskState(taskID int64, state model.TaskState) {
	select {
	case s.TaskStateChan <- model.TaskStateEvent{TaskID: taskID, State: state}:
	default:
	}
}

//////////////////////
// РЕСТОР
//////////////////////

func (s *AppService) RunFileRestore(
	backupPath string,
	targetDir string,
	overwrite bool,
) error {
	return s.RestoreSvc.RestoreFile(
		backupPath,
		targetDir,
		overwrite,
	)
}

func (s *AppService) RunFileRestoreWithChecksum(
	backupID int64,
	targetDir string,
	overwrite bool,
) error {

	backup, err := s.BackupRepo.GetByID(backupID)
	if err != nil {
		return err
	}

	if backup == nil {
		return fmt.Errorf("Копия %d не найдена", backupID)
	}

	return s.RestoreSvc.RestoreFileWithChecksum(
		backup.TargetPath,
		targetDir,
		backup.Checksum,
		overwrite,
	)
}

func (s *AppService) RunFolderRestore(
	backupDir string,
	targetDir string,
	overwrite bool,
) error {

	mode := RestoreToNewFolder
	if overwrite {
		mode = RestoreOverwrite
	}

	return s.RestoreSvc.RestoreFolder(backupDir, targetDir, mode)
}

func (s *AppService) RunFolderRestoreWithChecksum(
	backupID int64,
	targetDir string,
	overwrite bool,
) error {

	backup, err := s.BackupRepo.GetByID(backupID)
	if err != nil {
		return err
	}

	if backup == nil {
		return fmt.Errorf("Копия %d не найдена", backupID)
	}

	mode := RestoreToNewFolder
	if overwrite {
		mode = RestoreOverwrite
	}

	return s.RestoreSvc.RestoreFolderWithChecksum(
		backup.TargetPath,
		targetDir,
		backup.Checksum,
		mode,
	)
}

//////////////////////
// Дашборд
//////////////////////

// Статистика для карточек
func (s *AppService) GetBackupStats() (total, errors, upcoming int, err error) {
	total, err = s.BackupRepo.CountAll()
	if err != nil {
		return
	}

	errors, err = s.BackupRepo.CountByStatus("ERROR")
	if err != nil {
		return
	}

	upcoming, err = s.TaskRepo.CountUpcoming(time.Now(), time.Now().Add(24*time.Hour))
	return
}

// Последние бэкапы
func (s *AppService) GetLastBackups(limit int) ([]model.Backup, error) {
	return s.BackupRepo.GetLast(limit)
}

// Ближайшие задачи
func (s *AppService) GetUpcomingTasks(limit int) ([]model.Task, error) {
	return s.TaskRepo.GetUpcoming(limit)
}

// Проверка на заполненность хранилища
func (s *AppService) IsStorageExceeded() bool {
	used, err := s.GetUsedBytes(s.Settings.BackupRootPath)
	if err != nil {
		return false
	}
	return used > s.Settings.MaxStorageBytes
}

//////////////////////
// Статистика на дашборде
//////////////////////

func (s *AppService) SuccessCount() int {
	total, errors, _, err := s.GetBackupStats()
	if err != nil {
		return 0
	}
	return total - errors
}

func (s *AppService) ErrorCount() int {
	_, errors, _, err := s.GetBackupStats()
	if err != nil {
		return 0
	}
	return errors
}

func (s *AppService) UpcomingCount() int {
	_, _, upcoming, err := s.GetBackupStats()
	if err != nil {
		return 0
	}
	return upcoming
}

//////////////////////
// Хранилища
//////////////////////

func (s *AppService) GetStorageUsedBytes() (int64, error) {
	settings, err := s.SettingsRepo.Get()
	if err != nil {
		return 0, err
	}

	if settings.BackupRootPath == "" {
		return 0, nil
	}

	return s.GetUsedBytes(settings.BackupRootPath)
}

func (s *AppService) GetUsedBytes(rootPath string) (int64, error) {
	var total int64

	err := filepath.Walk(rootPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})

	return total, err
}

//////////////////////
// Логи
//////////////////////

func (s *AppService) GetBackupHistory(limit int, statusFilter string, dateFilter *time.Time, id *int64) ([]model.BackupWithTask, error) {
	return s.BackupRepo.GetHistory(limit, statusFilter, dateFilter, id)
}
