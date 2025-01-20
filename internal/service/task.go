package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/sev"
)

type TaskService struct {
	Sev              *sev.Sev
	TaskRepository   *repository.Task
	WebhookService   *WebhookService
	PresetService    *PresetService
	WebsocketService *WebsocketService
}

func (s *TaskService) ListTasks() (*[]model.Task, error) {
	return s.TaskRepository.List()
}

func (s *TaskService) GetTaskById(uuid string) (*model.Task, error) {
	return s.TaskRepository.First(uuid)
}

func (s *TaskService) GetTasksByBatchId(uuid string) (*[]model.Task, error) {
	return s.TaskRepository.ByBatchId(uuid)
}

func (s *TaskService) UpdateTask(task *model.Task) (*model.Task, error) {
	task, err := s.TaskRepository.UpdateTask(task)
	s.WebsocketService.Broadcast(TASK_UPDATED, task.ToDto())
	s.Sev.Metrics().Gauge("task.updated").Inc()
	s.WebhookService.Fire(dto.TASK_UPDATED, task.ToDto())
	return task, err
}

func (s *TaskService) DeleteTask(uuid string) error {
	w, err := s.TaskRepository.First(uuid)
	if err != nil {
		return err
	}

	if w.Uuid == "" {
		return errors.New("task for given uuid not found")
	}

	if w.Status == dto.RUNNING {
		return errors.New("running tasks can not be deleted, cancel first")
	}

	err = s.TaskRepository.Delete(w)
	if err != nil {
		s.Sev.Logger().Warnf("failed to delete task (uuid: %s): %+v", w.Uuid, err)
		return err
	}

	s.Sev.Logger().Infof("deleted task (uuid: %s)", w.Uuid)

	s.Sev.Metrics().Gauge("task.deleted").Inc()
	s.WebhookService.Fire(dto.TASK_DELETED, w.ToDto())
	s.WebsocketService.Broadcast(TASK_DELETED, w.ToDto())

	return nil
}

func (s *TaskService) CancelTask(uuid string) (*model.Task, error) {
	t, err := s.TaskRepository.First(uuid)
	if err != nil {
		return nil, err
	}

	if t.Status != dto.QUEUED {
		return nil, errors.New("failed to cancel job, not in status 'queue'")
	}

	t.Progress = 100
	t.Status = dto.DONE_CANCELED
	task, err := s.TaskRepository.UpdateTask(t)
	if err != nil {
		return nil, err
	}

	s.Sev.Metrics().Gauge("task.updated").Inc()
	s.WebhookService.Fire(dto.TASK_UPDATED, task.ToDto())
	s.WebsocketService.Broadcast(TASK_UPDATED, task.ToDto())

	return task, err
}

func (s *TaskService) NewTask(task *dto.NewTask, batch string) (*model.Task, error) {
	if task.Preset != "" {
		preset, err := s.PresetService.FindByName(task.Preset)
		if err != nil {
			return nil, err
		}
		task.Command = preset.Command
		if task.Priority == 0 {
			task.Priority = preset.Priority
		}
		if preset.PreProcessing != nil && task.PreProcessing == nil {
			task.PreProcessing = &dto.NewPrePostProcessing{ScriptPath: preset.PreProcessing.ScriptPath.Raw, SidecarPath: preset.PreProcessing.ScriptPath.Raw}
		}
		if preset.PostProcessing != nil && task.PostProcessing == nil {
			task.PostProcessing = &dto.NewPrePostProcessing{ScriptPath: preset.PostProcessing.ScriptPath.Raw, SidecarPath: preset.PostProcessing.ScriptPath.Raw}
		}
	}
	t, err := s.TaskRepository.Create(task, batch, "api")

	s.Sev.Metrics().Gauge("task.created").Inc()
	s.WebhookService.Fire(dto.TASK_CREATED, t.ToDto())
	s.WebsocketService.Broadcast(TASK_CREATED, t.ToDto())

	s.Sev.Logger().Infof("new task added to queue (uuid: %s)", t.Uuid)
	return t, err
}

func (s *TaskService) NewTasks(tasks *[]dto.NewTask) (*[]model.Task, error) {
	batch := uuid.NewString()
	newTasks := []model.Task{}
	for _, task := range *tasks {
		t, err := s.NewTask(&task, batch)
		if err != nil {
			return nil, err
		}

		newTasks = append(newTasks, *t)
	}

	// Transform each task to its DTO
	var taskDTOs = []dto.Task{}
	for _, task := range newTasks {
		taskDTOs = append(taskDTOs, *task.ToDto())
	}

	s.Sev.Metrics().Gauge("batch.created").Inc()
	s.WebhookService.Fire(dto.BATCH_CREATED, taskDTOs)
	return &newTasks, nil
}
