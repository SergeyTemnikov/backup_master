package service

import (
	"backup_master/internal/model"
	"backup_master/internal/repository"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type AppService struct {
	DB           *sql.DB
	TaskRepo     *repository.TaskRepository
	BackupRepo   *repository.BackupRepository
	SettingsRepo *repository.SettingsRepository

	Settings *model.AppSettings

	BackupSvc  *BackupService
	RestoreSvc *RestoreService

	Progress chan *model.BackupProgress

	Scheduler *Scheduler
}

func NewAppService(dbPath string) (*AppService, error) {
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

	svc := &AppService{
		DB:           db,
		TaskRepo:     repository.NewTaskRepository(db),
		BackupRepo:   repository.NewBackupRepository(db),
		SettingsRepo: settingsRepo,
		Settings:     settings,

		BackupSvc:  NewBackupService(),
		RestoreSvc: NewRestoreService(),

		Progress: make(chan *model.BackupProgress, 16),
	}

	svc.Scheduler = NewScheduler(svc)

	return svc, nil
}

//////////////////////
// Заглушка
//////////////////////

// EnsureDemoData добавляет тестовые данные, если БД пустая
func (s *AppService) EnsureDemoData() error {
	count, err := s.BackupRepo.CountAll()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := s.seedTasks(); err != nil {
		return err
	}
	if err := s.seedBackups(); err != nil {
		return err
	}

	return nil
}

func (s *AppService) seedTasks() error {
	tasks := []model.Task{
		{
			Name:       "Документы",
			SourcePath: "/home/user/Documents",
			Schedule:   "Каждый день в 21:00",
			Enabled:    true,
			CreatedAt:  time.Now(),
		},
		{
			Name:       "Фото",
			SourcePath: "/home/user/Pictures",
			Schedule:   "Каждый день в 23:00",
			Enabled:    true,
			CreatedAt:  time.Now(),
		},
	}

	for _, t := range tasks {
		if err := s.TaskRepo.Create(&t); err != nil {
			return err
		}
	}
	return nil
}

func (s *AppService) seedBackups() error {
	now := time.Now()

	backups := []model.Backup{
		{
			TaskID:     1,
			Status:     "OK",
			SizeBytes:  2 * 1024 * 1024 * 1024,
			StartedAt:  now.Add(-2 * time.Hour),
			FinishedAt: now.Add(-1*time.Hour + -30*time.Minute),
		},
		{
			TaskID:       2,
			Status:       "ERROR",
			SizeBytes:    0,
			ErrorMessage: ptr("Недостаточно места"),
			StartedAt:    now.Add(-5 * time.Hour),
			FinishedAt:   now.Add(-5*time.Hour + 10*time.Minute),
		},
	}

	for _, b := range backups {
		_, err := s.DB.Exec(`
			INSERT INTO backups (task_id, status, size_bytes, error_message, started_at, finished_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`,
			b.TaskID, b.Status, b.SizeBytes, b.ErrorMessage, b.StartedAt, b.FinishedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ptr(s string) *string {
	return &s
}

//////////////////////
// РУЧНОЙ БЭКАП
//////////////////////

func (s *AppService) RunManualBackup(srcFile, dstFolder string) error {
	size, err := s.BackupSvc.BackupFile(srcFile, dstFolder)
	return s.saveManualBackup(size, err)
}

func (s *AppService) RunManualFolderBackup(srcFile, dstFolder string) error {
	size, err := s.BackupSvc.BackupFolder(srcFile, dstFolder)
	return s.saveManualBackup(size, err)
}

func (s *AppService) saveManualBackup(size int64, err error) error {
	started := time.Now()

	status := "OK"
	var errMsg *string

	if err != nil {
		status = "ERROR"
		msg := err.Error()
		errMsg = &msg
	}

	_, dbErr := s.DB.Exec(`
		INSERT INTO backups (
			task_id,
			status,
			size_bytes,
			error_message,
			started_at,
			finished_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		nil,
		status,
		size,
		errMsg,
		started,
		time.Now(),
	)

	if dbErr != nil {
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
	started := time.Now()

	s.Progress <- &model.BackupProgress{
		TaskID:  task.ID,
		Percent: 0,
		Message: "Запуск задачи",
	}

	if err := s.CheckStorageLimit(); err != nil {
		s.sendTaskError(task.ID, err)
		return
	}

	var (
		size int64
		err  error
	)

	fmt.Printf(
		"DEBUG task %d sourceType=%q\n",
		task.ID,
		task.SourceType,
	)

	switch task.SourceType {
	case "file":
		size, err = s.BackupSvc.BackupFile(
			task.SourcePath,
			s.Settings.BackupRootPath,
		)

	case "folder":
		size, err = s.BackupSvc.BackupFolder(
			task.SourcePath,
			s.Settings.BackupRootPath,
		)

	default:
		s.sendTaskError(task.ID, fmt.Errorf("неизвестный тип источника"))
		return
	}

	status := "OK"
	var errMsg *string

	if err != nil {
		status = "ERROR"
		msg := err.Error()
		errMsg = &msg
	}

	_, dbErr := s.DB.Exec(`
		INSERT INTO backups (
			task_id,
			status,
			size_bytes,
			error_message,
			started_at,
			finished_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		task.ID,
		status,
		size,
		errMsg,
		started,
		time.Now(),
	)

	if dbErr != nil {
		s.sendTaskError(task.ID, dbErr)
		return
	}

	s.Progress <- &model.BackupProgress{
		TaskID:  task.ID,
		Percent: 100,
		Message: "Задача завершена",
	}
}

func (s *AppService) sendTaskError(taskID int64, err error) {
	s.Progress <- &model.BackupProgress{
		TaskID:  taskID,
		Message: err.Error(),
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
// Авто бэкап
// ////////////////////
func (s *AppService) RunTask(task model.Task) {
	go s.runTask(task)
}

// ////////////////////
// Работа с тасками
// ////////////////////

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

//////////////////////
// DASHBOARD
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

// ====== DASHBOARD SHORT METHODS ======

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

// ====== STORAGES ======

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
