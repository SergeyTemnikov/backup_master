package repository

import (
	"database/sql"
	"errors"
	"testing"

	"backup_master/internal/model"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

type SettingsRepoTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo *SettingsRepository
}

func (s *SettingsRepoTestSuite) SetupTest() {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	s.Require().NoError(err)
	s.db = db

	s.createSchema()
	s.seedDefaultSettings()

	s.repo = NewSettingsRepository(db)
}

func (s *SettingsRepoTestSuite) TearDownTest() {
	s.db.Close()
}

func (s *SettingsRepoTestSuite) createSchema() {
	_, err := s.db.Exec(`
		CREATE TABLE settings (
			id INTEGER PRIMARY KEY,
			backup_root_path TEXT NOT NULL,
			max_storage_bytes INTEGER NOT NULL,
			theme_mode TEXT NOT NULL
		)
	`)
	s.Require().NoError(err)
}

func (s *SettingsRepoTestSuite) seedDefaultSettings() {
	_, err := s.db.Exec(`
		INSERT INTO settings (id, backup_root_path, max_storage_bytes, theme_mode)
		VALUES (1, '/default/backup', 1073741824, 'light')
	`)
	s.Require().NoError(err)
}

func (s *SettingsRepoTestSuite) TestGet_Success() {
	settings, err := s.repo.Get()

	s.Require().NoError(err)
	s.Equal("/default/backup", settings.BackupRootPath)
	s.Equal(int64(1073741824), settings.MaxStorageBytes)
	s.Equal("light", settings.ThemeMode)
}

func (s *SettingsRepoTestSuite) TestGet_NotFound() {
	_, err := s.db.Exec("DELETE FROM settings WHERE id = 1")
	s.Require().NoError(err)

	settings, err := s.repo.Get()

	s.True(errors.Is(err, sql.ErrNoRows))
	s.Nil(settings)
}

func (s *SettingsRepoTestSuite) TestSave_Success() {
	newSettings := &model.AppSettings{
		BackupRootPath:  "/new/path",
		MaxStorageBytes: 5000000,
		ThemeMode:       "dark",
	}

	err := s.repo.Save(newSettings)
	s.Require().NoError(err)

	updated, err := s.repo.Get()
	s.Require().NoError(err)

	s.Equal("/new/path", updated.BackupRootPath)
	s.Equal(int64(5000000), updated.MaxStorageBytes)
	s.Equal("dark", updated.ThemeMode)
}

func (s *SettingsRepoTestSuite) TestUpdateBackupSettings() {
	err := s.repo.UpdateBackupSettings("/updated/backup", 999999)
	s.Require().NoError(err)

	settings, err := s.repo.Get()
	s.Require().NoError(err)

	s.Equal("/updated/backup", settings.BackupRootPath)
	s.Equal(int64(999999), settings.MaxStorageBytes)
	s.Equal("light", settings.ThemeMode)
}

func (s *SettingsRepoTestSuite) TestUpdateTheme() {
	err := s.repo.UpdateTheme("dark")
	s.Require().NoError(err)

	settings, err := s.repo.Get()
	s.Require().NoError(err)

	s.Equal("dark", settings.ThemeMode)
	s.Equal("/default/backup", settings.BackupRootPath)
}

func (s *SettingsRepoTestSuite) TestUpdateBackupSettings_EmptyPath() {
	err := s.repo.UpdateBackupSettings("", 0)
	s.Require().NoError(err)

	settings, err := s.repo.Get()
	s.Require().NoError(err)
	s.Equal("", settings.BackupRootPath)
	s.Equal(int64(0), settings.MaxStorageBytes)
}

func TestSettingsRepoSuite(t *testing.T) {
	suite.Run(t, new(SettingsRepoTestSuite))
}
