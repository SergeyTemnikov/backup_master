package service

import (
	"log"
	"sync"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
	svc  *AppService
	mu   sync.Mutex
}

func NewScheduler(svc *AppService) *Scheduler {
	return &Scheduler{
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithLogger(cron.VerbosePrintfLogger(
				log.New(log.Writer(), "[cron] ", log.LstdFlags),
			)),
		),
		svc: svc,
	}
}

func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loadTasks()
	s.cron.Start()

	log.Println("scheduler started")
	return nil
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *Scheduler) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("scheduler reload")

	s.cron.Stop()
	s.cron = cron.New(
		cron.WithSeconds(),
		cron.WithLogger(cron.VerbosePrintfLogger(
			log.New(log.Writer(), "[cron] ", log.LstdFlags),
		)),
	)

	s.loadTasks()
	s.cron.Start()
}

func (s *Scheduler) loadTasks() {
	tasks, err := s.svc.TaskRepo.GetEnabled()
	if err != nil {
		log.Println("scheduler load error:", err)
		return
	}

	for _, task := range tasks {
		task := task // важно

		_, err := s.cron.AddFunc(task.Schedule, func() {
			s.svc.runTask(task)
		})

		if err != nil {
			log.Println("cron task error:", err, task.Name)
		}
	}
}
