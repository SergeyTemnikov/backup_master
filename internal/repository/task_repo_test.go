package repository

import (
	"database/sql"
	"testing"
	"time"

	"backup_master/internal/model"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

type TaskRepoTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo *TaskRepository
}

func (s *TaskRepoTestSuite) SetupTest() {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	s.Require().NoError(err)
	s.db = db

	s.createSchema()
	s.repo = NewTaskRepository(db)
}

func (s *TaskRepoTestSuite) TearDownTest() {
	s.db.Close()
}

func (s *TaskRepoTestSuite) createSchema() {
	_, err := s.db.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			source_path TEXT NOT NULL,
			source_type TEXT NOT NULL,
			schedule TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL
		)
	`)
	s.Require().NoError(err)
}

func (s *TaskRepoTestSuite) createTestTask(name string, enabled bool) int64 {
	var id int64
	err := s.db.QueryRow(`
		INSERT INTO tasks (name, source_path, source_type, schedule, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
		RETURNING id
	`, name, "/data", "folder", "0 0 * * *", enabled).Scan(&id)

	if err != nil {
		res, err := s.db.Exec(`
			INSERT INTO tasks (name, source_path, source_type, schedule, enabled, created_at)
			VALUES (?, ?, ?, ?, ?, datetime('now'))
		`, name, "/data", "folder", "0 0 * * *", enabled)
		s.Require().NoError(err)
		id, _ = res.LastInsertId()
	}
	return id
}

func (s *TaskRepoTestSuite) TestCreate_Success() {
	createdAt := time.Now().Truncate(time.Second)

	task := &model.Task{
		Name:       "Daily Backup",
		SourcePath: "/home/user/data",
		SourceType: "folder",
		Schedule:   "0 0 * * *",
		Enabled:    true,
		CreatedAt:  createdAt,
	}

	err := s.repo.Create(task)
	s.Require().NoError(err)

	var count int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE name = ?`, task.Name).Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count, "Запись должна появиться в БД")

	var saved model.Task
	err = s.db.QueryRow(`
		SELECT id, name, source_path, source_type, schedule, enabled, created_at
		FROM tasks WHERE name = ?
	`, task.Name).Scan(
		&saved.ID,
		&saved.Name,
		&saved.SourcePath,
		&saved.SourceType,
		&saved.Schedule,
		&saved.Enabled,
		&saved.CreatedAt,
	)
	s.Require().NoError(err)

	s.Equal(task.Name, saved.Name)
	s.Equal(task.SourcePath, saved.SourcePath)
	s.Equal(task.SourceType, saved.SourceType)
	s.Equal(task.Schedule, saved.Schedule)
	s.Equal(task.Enabled, saved.Enabled)

	s.WithinDuration(task.CreatedAt, saved.CreatedAt, 2*time.Second)
}
func (s *TaskRepoTestSuite) TestCreate_ValidateFields() {
	task := &model.Task{
		Name:       "",
		SourcePath: "/data",
		SourceType: "folder",
		Schedule:   "0 0 * * *",
		Enabled:    true,
		CreatedAt:  time.Now(),
	}

	err := s.repo.Create(task)
	s.NoError(err)
}

func (s *TaskRepoTestSuite) TestGetAll_Empty() {
	tasks, err := s.repo.GetAll()
	s.Require().NoError(err)
	s.Len(tasks, 0)
}

func (s *TaskRepoTestSuite) TestGetAll_WithTasks() {
	s.createTestTask("Task 1", true)
	s.createTestTask("Task 2", false)
	s.createTestTask("Task 3", true)

	tasks, err := s.repo.GetAll()
	s.Require().NoError(err)
	s.Len(tasks, 3)

	s.Equal("Task 1", tasks[0].Name)
	s.Equal("Task 2", tasks[1].Name)
	s.Equal("Task 3", tasks[2].Name)
}

func (s *TaskRepoTestSuite) TestGetUpcoming_OnlyEnabled() {
	s.createTestTask("Enabled Task", true)
	s.createTestTask("Disabled Task", false)

	tasks, err := s.repo.GetUpcoming(10)
	s.Require().NoError(err)
	s.Len(tasks, 1)
	s.Equal("Enabled Task", tasks[0].Name)
}

func (s *TaskRepoTestSuite) TestGetUpcoming_Limit() {
	for i := 0; i < 5; i++ {
		s.createTestTask("Task", true)
	}

	tasks, err := s.repo.GetUpcoming(2)
	s.Require().NoError(err)
	s.Len(tasks, 2)
}

func (s *TaskRepoTestSuite) TestCountUpcoming_EnabledOnly() {
	s.createTestTask("Enabled", true)
	s.createTestTask("Enabled 2", true)
	s.createTestTask("Disabled", false)

	count, err := s.repo.CountUpcoming(time.Now(), time.Now().Add(24*time.Hour))
	s.Require().NoError(err)
	s.Equal(2, count)
}

func (s *TaskRepoTestSuite) TestCountUpcoming_WithDateRange() {
	count, err := s.repo.CountUpcoming(
		time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2050, 1, 2, 0, 0, 0, 0, time.UTC),
	)
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *TaskRepoTestSuite) TestUpdate_Success() {
	id := s.createTestTask("Old Name", true)

	task := &model.Task{
		ID:         id,
		Name:       "New Name",
		SourcePath: "/new/path",
		SourceType: "folder",
		Schedule:   "0 6 * * *",
		Enabled:    false,
		CreatedAt:  time.Now(),
	}

	err := s.repo.Update(task)
	s.Require().NoError(err)

	updated, err := s.repo.GetAll()
	s.Require().NoError(err)

	var found *model.Task
	for _, t := range updated {
		if t.ID == id {
			found = &t
			break
		}
	}

	s.Require().NotNil(found)
	s.Equal("New Name", found.Name)
	s.Equal("/new/path", found.SourcePath)
	s.Equal(false, found.Enabled)
}

func (s *TaskRepoTestSuite) TestUpdate_NotFound() {
	task := &model.Task{
		ID:         999,
		Name:       "Test",
		SourcePath: "/data",
		SourceType: "folder",
		Schedule:   "0 0 * * *",
		Enabled:    true,
		CreatedAt:  time.Now(),
	}

	err := s.repo.Update(task)
	s.Require().NoError(err)
}

func (s *TaskRepoTestSuite) TestDelete_Success() {
	id := s.createTestTask("To Delete", true)

	err := s.repo.Delete(id)
	s.Require().NoError(err)

	tasks, _ := s.repo.GetAll()
	s.Len(tasks, 0)
}

func (s *TaskRepoTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(999)
	s.Require().NoError(err)
}

func (s *TaskRepoTestSuite) TestSetEnabled_Disable() {
	id := s.createTestTask("Task", true)

	err := s.repo.SetEnabled(id, false)
	s.Require().NoError(err)

	tasks, _ := s.repo.GetEnabled()
	s.Len(tasks, 0)
}

func (s *TaskRepoTestSuite) TestSetEnabled_Enable() {
	id := s.createTestTask("Task", false)

	err := s.repo.SetEnabled(id, true)
	s.Require().NoError(err)

	tasks, err := s.repo.GetEnabled()
	s.Require().NoError(err)
	s.Len(tasks, 1)
	s.Equal(id, tasks[0].ID)
}

func (s *TaskRepoTestSuite) TestGetEnabled_Filter() {
	s.createTestTask("Enabled 1", true)
	s.createTestTask("Disabled", false)
	s.createTestTask("Enabled 2", true)

	tasks, err := s.repo.GetEnabled()
	s.Require().NoError(err)
	s.Len(tasks, 2)

	for _, t := range tasks {
		s.True(t.Enabled)
	}
}

func TestTaskRepoSuite(t *testing.T) {
	suite.Run(t, new(TaskRepoTestSuite))
}
