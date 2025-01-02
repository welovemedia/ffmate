package service

import (
	"errors"

	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/sev"
)

type TaskService struct {
	Sev            *sev.Sev
	TaskRepository *repository.Task
	WebhookService *WebhookService
}

func (s *TaskService) ListTasks() (*[]model.Task, error) {
	return s.TaskRepository.List()
}

func (s *TaskService) GetTaskById(uuid string) (*model.Task, error) {
	return s.TaskRepository.First(uuid)
}

func (s *TaskService) CancelTask(uuid string) (*model.Task, error) {
	t, err := s.TaskRepository.First(uuid)
	if err != nil {
		return nil, err
	}

	if t.Status != dto.QUEUED {
		return nil, errors.New("failed to cancel job, not in status 'queue'")
	}

	task, err := s.TaskRepository.SetTaskStatus(t, dto.DONE_CANCELED)
	if err != nil {
		return nil, err
	}

	s.Sev.Metrics().Gauge("task.status.updated").Inc()
	s.WebhookService.Fire(dto.TASK_STATUS_UPDATED, task.ToDto())

	return task, err
}

func (s *TaskService) NewTask(task *dto.NewTask) (*model.Task, error) {
	t, err := s.TaskRepository.Create(task.Command, task.InputFile, task.OutputFile)

	s.Sev.Metrics().Gauge("task.created").Inc()
	s.WebhookService.Fire(dto.TASK_CREATED, t.ToDto())

	s.Sev.Logger().Infof("new task added to queue (uuid: %s)", t.Uuid)
	return t, err
}
