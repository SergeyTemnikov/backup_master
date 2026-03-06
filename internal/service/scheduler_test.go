package service

import (
	"backup_master/internal/model"
	"backup_master/internal/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheduler_LoadTasks_OnlyEnabled(t *testing.T) {
	mockTaskRepo := new(mocks.TaskRepositoryInterface)
	mockAppSvc := &AppService{TaskRepo: mockTaskRepo}

	mockTaskRepo.On("GetEnabled").Return([]model.Task{
		{ID: 1, Name: "Daily", Schedule: "0 0 0 * * *", Enabled: true},
	}, nil).Once()

	scheduler := NewScheduler(mockAppSvc)
	scheduler.loadTasks()

	entries := scheduler.cron.Entries()
	assert.Equal(t, 1, len(entries), "Должна быть зарегистрирована только одна задача")

	mockTaskRepo.AssertExpectations(t)
}

func TestScheduler_Reload_StopsOldCron(t *testing.T) {
	mockTaskRepo := new(mocks.TaskRepositoryInterface)
	mockAppSvc := &AppService{TaskRepo: mockTaskRepo}

	mockTaskRepo.On("GetEnabled").Return([]model.Task{
		{ID: 1, Name: "Daily", Schedule: "0 0 0 * * *", Enabled: true},
		{ID: 2, Name: "Hourly", Schedule: "0 0 * * * *", Enabled: true},
	}, nil).Once()

	scheduler := NewScheduler(mockAppSvc)
	scheduler.loadTasks()

	initialCount := len(scheduler.cron.Entries())
	assert.Equal(t, 2, initialCount, "После первой загрузки должно быть 2 задачи")

	mockTaskRepo.On("GetEnabled").Return([]model.Task{
		{ID: 3, Name: "Weekly", Schedule: "0 0 0 * * 0", Enabled: true},
	}, nil).Once()

	scheduler.Reload()

	afterReloadCount := len(scheduler.cron.Entries())
	assert.Equal(t, 1, afterReloadCount, "После Reload должна быть только одна новая задача")

	assert.NotEqual(t, initialCount, afterReloadCount, "Список задач должен измениться после Reload")

	mockTaskRepo.AssertNumberOfCalls(t, "GetEnabled", 2)
	mockTaskRepo.AssertExpectations(t)
}

func TestScheduler_Reload_UpdatesTasks(t *testing.T) {
	mockTaskRepo := new(mocks.TaskRepositoryInterface)
	mockAppSvc := &AppService{TaskRepo: mockTaskRepo}

	mockTaskRepo.On("GetEnabled").Return([]model.Task{
		{ID: 1, Name: "Daily", Schedule: "0 0 0 * * *", Enabled: true},
		{ID: 2, Name: "Hourly", Schedule: "0 0 * * * *", Enabled: true},
	}, nil).Once()

	scheduler := NewScheduler(mockAppSvc)
	scheduler.loadTasks()

	entries := scheduler.cron.Entries()
	assert.Equal(t, 2, len(entries), "После первой загрузки должно быть 2 задачи")

	mockTaskRepo.On("GetEnabled").Return([]model.Task{
		{ID: 3, Name: "Weekly", Schedule: "0 0 0 * * 0", Enabled: true},
	}, nil).Once()

	scheduler.Reload()

	afterReload := scheduler.cron.Entries()
	assert.Equal(t, 1, len(afterReload), "После Reload должна остаться только одна задача")

	mockTaskRepo.AssertNumberOfCalls(t, "GetEnabled", 2)
	mockTaskRepo.AssertExpectations(t)
}

func TestScheduler_Reload_Idempotent(t *testing.T) {
	mockTaskRepo := new(mocks.TaskRepositoryInterface)
	mockAppSvc := &AppService{TaskRepo: mockTaskRepo}

	mockTaskRepo.On("GetEnabled").Return([]model.Task{
		{ID: 1, Name: "Stable", Schedule: "0 0 * * * *", Enabled: true},
	}, nil)

	scheduler := NewScheduler(mockAppSvc)
	scheduler.loadTasks()

	for i := 0; i < 3; i++ {
		scheduler.Reload()
		count := len(scheduler.cron.Entries())
		assert.Equal(t, 1, count, "После Reload #%d должна быть ровно 1 задача", i+1)
	}

	mockTaskRepo.AssertNumberOfCalls(t, "GetEnabled", 4)
	mockTaskRepo.AssertExpectations(t)
}
