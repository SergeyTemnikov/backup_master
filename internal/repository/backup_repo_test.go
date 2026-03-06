package repository

import (
	"database/sql"
	"testing"
	"time"

	"backup_master/internal/model"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

type BackupRepoTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo *BackupRepository
}

func (s *BackupRepoTestSuite) SetupTest() {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	s.Require().NoError(err)
	s.db = db

	s.createSchema()

	s.repo = NewBackupRepository(db)
}

func (s *BackupRepoTestSuite) TearDownTest() {
	s.db.Close()
}

func (s *BackupRepoTestSuite) createSchema() {
	_, err := s.db.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			source_path TEXT,
			source_type TEXT,
			schedule TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME
		);
		CREATE TABLE backups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER,
			status TEXT,
			size_bytes INTEGER,
			source_path TEXT,
			source_type TEXT,
			target_path TEXT,
			error_message TEXT,
			checksum TEXT,
			started_at DATETIME,
			finished_at DATETIME
		);
	`)
	s.Require().NoError(err)
}

func (s *BackupRepoTestSuite) TestCreateAndGetByID() {
	now := time.Now().Truncate(time.Second)

	uniquePath := "/backup/test_" + time.Now().Format("150405") + ".zip"

	backup := &model.Backup{
		TaskID:     int64Ptr(1),
		Status:     "completed",
		SizeBytes:  1024,
		SourcePath: "/data",
		SourceType: "folder",
		TargetPath: uniquePath,
		StartedAt:  now,
		FinishedAt: now,
		Checksum:   "abc123",
	}

	err := s.repo.Create(backup)
	s.Require().NoError(err)

	var found model.Backup
	err = s.db.QueryRow(`
		SELECT id, task_id, status, size_bytes, source_path, source_type,
		       target_path, error_message, checksum, started_at, finished_at
		FROM backups WHERE target_path = ?
	`, uniquePath).Scan(
		&found.ID,
		&found.TaskID,
		&found.Status,
		&found.SizeBytes,
		&found.SourcePath,
		&found.SourceType,
		&found.TargetPath,
		&found.ErrorMessage,
		&found.Checksum,
		&found.StartedAt,
		&found.FinishedAt,
	)
	s.Require().NoError(err)

	s.Equal(backup.Status, found.Status)
	s.Equal(backup.SizeBytes, found.SizeBytes)
	s.Equal(backup.SourcePath, found.SourcePath)
	s.Equal(backup.SourceType, found.SourceType)
	s.Equal(backup.TargetPath, found.TargetPath)
	s.Equal(backup.Checksum, found.Checksum)

	s.WithinDuration(backup.StartedAt, found.StartedAt, 2*time.Second)
	s.WithinDuration(backup.FinishedAt, found.FinishedAt, 2*time.Second)

	s.Greater(found.ID, int64(0), "БД должна присвоить ID записи")
}

func (s *BackupRepoTestSuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(999)
	s.Error(err)
}

func (s *BackupRepoTestSuite) TestCounters() {
	s.createTestBackupFull("OK")
	s.createTestBackupFull("OK")
	s.createTestBackupFull("ERROR")

	total, err := s.repo.CountAll()
	s.Require().NoError(err)
	s.Equal(3, total)

	completed, err := s.repo.CountByStatus("OK")
	s.Require().NoError(err)
	s.Equal(2, completed)
}

func (s *BackupRepoTestSuite) TestGetLast_Limit() {
	for i := 0; i < 5; i++ {
		s.createTestBackupFull("OK")
	}

	backups, err := s.repo.GetLast(2)
	s.Require().NoError(err)
	s.Len(backups, 2)
}

func (s *BackupRepoTestSuite) TestGetHistory_Filtering() {
	taskID := s.createTestTask("Daily Task")

	_, err := s.db.Exec(`
		INSERT INTO backups (
			task_id, status, size_bytes, source_path, source_type,
			target_path, error_message, checksum, started_at, finished_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		taskID,
		"OK",
		int64(2048),
		"/home/user/data",
		"folder",
		"/backup/daily_20231001.zip",
		"",
		"sha256:abc123",
		"2023-10-01 10:00:00",
		"2023-10-01 10:05:00",
	)
	s.Require().NoError(err)

	_, err = s.db.Exec(`
		INSERT INTO backups (
			task_id, status, size_bytes, source_path, source_type,
			target_path, error_message, checksum, started_at, finished_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		nil,
		"OK",
		int64(512),
		"/manual/important",
		"file",
		"/backup/manual.zip",
		"",
		"sha256:def456",
		"2023-10-02 10:00:00",
		"2023-10-02 10:02:00",
	)
	s.Require().NoError(err)

	res, err := s.repo.GetHistory(10, "OK", nil, nil)
	s.Require().NoError(err)
	s.Len(res, 2)

	manualID := int64(-1)
	res, err = s.repo.GetHistory(10, "", nil, &manualID)
	s.Require().NoError(err)
	s.Len(res, 1)
	s.Equal("Ручной бэкап", res[0].TaskName)
	s.Equal("/manual/important", res[0].SourcePath)
}

func (s *BackupRepoTestSuite) createTestBackup(status string) {
	_, err := s.db.Exec(`
		INSERT INTO backups (
			status, size_bytes, source_path, source_type, 
			target_path, started_at, finished_at
		) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`,
		status,
		int64(1024),
		"/data",
		"folder",
		"/backup/test.zip",
	)
	s.Require().NoError(err)
}

func (s *BackupRepoTestSuite) createTestBackupFull(status string) {
	_, err := s.db.Exec(`
		INSERT INTO backups (
			task_id, status, size_bytes, source_path, source_type,
			target_path, error_message, checksum, started_at, finished_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`,
		int64(1),
		status,
		int64(1024),
		"/data",
		"folder",
		"/backup/test.zip",
		"",
		"abc123",
	)
	s.Require().NoError(err)
}
func (s *BackupRepoTestSuite) createTestTask(name string) int64 {
	var id int64
	err := s.db.QueryRow(`
		INSERT INTO tasks (name, source_path, source_type, schedule, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
		RETURNING id
	`, name, "/data", "folder", "0 0 * * *", true).Scan(&id)

	if err != nil {
		res, err := s.db.Exec(`
			INSERT INTO tasks (name, source_path, source_type, schedule, enabled, created_at)
			VALUES (?, ?, ?, ?, ?, datetime('now'))
		`, name, "/data", "folder", "0 0 * * *", true)
		s.Require().NoError(err)
		id, _ = res.LastInsertId()
	}
	return id
}
func TestBackupRepoSuite(t *testing.T) {
	suite.Run(t, new(BackupRepoTestSuite))
}

func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
