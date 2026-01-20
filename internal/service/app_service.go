package service

import (
	"backup_master/internal/model"
	"backup_master/internal/repository"
	"database/sql"
	"time"
)

type AppService struct {
	DB           *sql.DB
	TaskRepo     *repository.TaskRepository
	BackupRepo   *repository.BackupRepository
	StorageRepo  *repository.StorageRepository
	SettingsRepo *repository.SettingsRepository

	Settings *model.AppSettings

	BackupSvc  *BackupService
	RestoreSvc *RestoreService
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

	return &AppService{
		DB:           db,
		TaskRepo:     repository.NewTaskRepository(db),
		BackupRepo:   repository.NewBackupRepository(db),
		StorageRepo:  repository.NewStorageRepository(db),
		SettingsRepo: settingsRepo,
		Settings:     settings,

		BackupSvc:  NewBackupService(),
		RestoreSvc: NewRestoreService(),
	}, nil
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

	if err := s.seedStorages(); err != nil {
		return err
	}
	if err := s.seedTasks(); err != nil {
		return err
	}
	if err := s.seedBackups(); err != nil {
		return err
	}

	return nil
}

func (s *AppService) seedStorages() error {
	storages := []model.Storage{
		{
			Name:      "Локальный диск C:",
			Path:      "/",
			MaxBytes:  500 * 1024 * 1024 * 1024, // 500 GB
			UsedBytes: 120 * 1024 * 1024 * 1024,
			CreatedAt: time.Now(),
		},
	}

	for _, st := range storages {
		_, err := s.DB.Exec(`
			INSERT INTO storages (name, path, max_bytes, used_bytes, created_at)
			VALUES (?, ?, ?, ?, ?)
		`,
			st.Name, st.Path, st.MaxBytes, st.UsedBytes, st.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AppService) seedTasks() error {
	tasks := []model.Task{
		{
			Name:       "Документы",
			SourcePath: "/home/user/Documents",
			StorageID:  1,
			Schedule:   "Каждый день в 21:00",
			Enabled:    true,
			CreatedAt:  time.Now(),
		},
		{
			Name:       "Фото",
			SourcePath: "/home/user/Pictures",
			StorageID:  1,
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
	started := time.Now()

	size, err := s.BackupSvc.BackupFile(srcFile, dstFolder)

	status := "OK"
	var errMsg *string

	if err != nil {
		status = "ERROR"
		msg := err.Error()
		errMsg = &msg
	}

	_, dbErr := s.DB.Exec(`
		INSERT INTO backups (task_id, status, size_bytes, error_message, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		nil, // task_id = NULL (ручной запуск)
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

func (s *AppService) RunFolderBackup(sourceDir, targetDir string) error {
	started := time.Now()

	size, err := s.BackupSvc.BackupFolder(sourceDir, targetDir)

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
		nil, // ручной запуск
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
// РУЧНОЙ РЕСТОР
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

// Хранилища
func (s *AppService) GetStorages() ([]model.Storage, error) {
	return s.StorageRepo.GetAll()
}

// Ближайшие задачи
func (s *AppService) GetUpcomingTasks(limit int) ([]model.Task, error) {
	return s.TaskRepo.GetUpcoming(limit)
}

// Проверка на заполненность хранилища
func (s *AppService) IsStorageExceeded() bool {
	used, err := s.StorageRepo.CalcDirSize(s.Settings.BackupRootPath)
	if err != nil {
		return false
	}
	return used > s.Settings.MaxStorageBytes
}

//////////////////////
// TASKS / PLANNER
//////////////////////

func (s *AppService) GetAllTasks() ([]model.Task, error) {
	return s.TaskRepo.GetAll()
}

func (s *AppService) CreateTask(task *model.Task) error {
	return s.TaskRepo.Create(task)
}

func (s *AppService) SetTaskEnabled(taskID int64, enabled bool) error {
	return s.TaskRepo.SetEnabled(taskID, enabled)
}

//////////////////////
// BACKUPS / RECOVERY
//////////////////////

func (s *AppService) GetAllBackups() ([]model.Backup, error) {
	return s.BackupRepo.GetAll()
}

// ====== DASHBOARD SHORT METHODS ======

func (s *AppService) SuccessCount() int {
	total, _, _, err := s.GetBackupStats()
	if err != nil {
		return 0
	}
	return total
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

	return s.StorageRepo.GetUsedBytes(settings.BackupRootPath)
}

// ====== BACKUPS ======

func (s *AppService) LastBackups() []model.Backup {
	backups, err := s.GetLastBackups(5)
	if err != nil {
		return nil
	}
	return backups
}
