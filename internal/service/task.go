package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/sev"
)

type taskSvc struct {
	service
	sev            *sev.Sev
	taskRepository *repository.Task
}

var taskUpdates = make(chan *model.Task, 100)

func (s *taskSvc) CountAllStatus(session bool) (queued, running, doneSuccessful, doneError, doneCanceled int, err error) {
	if session {
		return s.taskRepository.CountAllStatus(s.sev.Session())
	}
	return s.taskRepository.CountAllStatus("")
}

func (s *taskSvc) GetTaskUpdates() chan *model.Task {
	return taskUpdates
}

func (s *taskSvc) ListTasks(page int, perPage int, status string) (*[]model.Task, int64, error) {
	return s.taskRepository.List(page, perPage, status)
}

func (s *taskSvc) GetTaskByUuid(uuid string) (*model.Task, error) {
	return s.taskRepository.First(uuid)
}

func (s *taskSvc) GetTasksByBatchId(uuid string, page int, perPage int) (*[]model.Task, int64, error) {
	return s.taskRepository.ByBatchId(uuid, page, perPage)
}

func (s *taskSvc) UpdateTask(task *model.Task) (*model.Task, error) {
	task, err := s.taskRepository.UpdateTask(task)
	WebsocketService().Broadcast(TASK_UPDATED, task.ToDto())
	s.sev.Metrics().Gauge("task.updated").Inc()
	WebhookService().Fire(dto.TASK_UPDATED, task.ToDto())

	if task.Batch != "" {
		switch task.Status {
		case dto.DONE_SUCCESSFUL, dto.DONE_ERROR, dto.DONE_CANCELED:
			c, _ := s.taskRepository.CountNonFinishedTasksByBatchId(task.Batch)
			if c == 0 {
				WebsocketService().Broadcast(BATCH_FINISHED, task.ToDto())
				s.sev.Metrics().Gauge("batch.finished").Inc()
				WebhookService().Fire(dto.BATCH_FINISHED, task.ToDto())
			}
		}
	}

	return task, err
}

func (s *taskSvc) DeleteTask(uuid string) error {
	w, err := s.taskRepository.First(uuid)
	if err != nil {
		return err
	}

	if w.Uuid == "" {
		return errors.New("task for given uuid not found")
	}

	if w.Status == dto.RUNNING {
		return errors.New("running tasks can not be deleted, cancel first")
	}

	err = s.taskRepository.Delete(w)
	if err != nil {
		s.sev.Logger().Warnf("failed to delete task (uuid: %s): %+v", w.Uuid, err)
		return err
	}

	s.sev.Logger().Infof("deleted task (uuid: %s)", w.Uuid)

	s.sev.Metrics().Gauge("task.deleted").Inc()
	WebhookService().Fire(dto.TASK_DELETED, w.ToDto())
	WebsocketService().Broadcast(TASK_DELETED, w.ToDto())

	return nil
}

func (s *taskSvc) RestartTask(uuid string) (*model.Task, error) {
	t, err := s.GetTaskByUuid(uuid)
	if err != nil {
		return nil, err
	}

	if t.Status == dto.QUEUED {
		return nil, errors.New("failed to restart task, task is already in status 'queue'")
	}

	t.Progress = 0
	t.StartedAt = 0
	t.FinishedAt = 0
	t.Error = ""
	t.Status = dto.QUEUED
	s.sev.Metrics().Gauge("task.restarted").Inc()
	return s.UpdateTask(t)
}

func (s *taskSvc) CancelTask(uuid string) (*model.Task, error) {
	t, err := s.GetTaskByUuid(uuid)
	if err != nil {
		return nil, err
	}

	if t.Status != dto.QUEUED && t.Status != dto.RUNNING {
		return nil, errors.New("failed to cancel task, task in unsupported state")
	}

	if t.Status == dto.RUNNING {
		taskUpdates <- t
	}

	t.Progress = 100
	t.Remaining = -1
	t.FinishedAt = time.Now().UnixMilli()
	t.Status = dto.DONE_CANCELED
	s.sev.Metrics().Gauge("task.canceled").Inc()
	return s.UpdateTask(t)
}

func (s *taskSvc) NewTask(task *dto.NewTask, batch string, source string) (*model.Task, error) {
	if task.Preset != "" {
		preset, err := PresetService().FindByUuid(task.Preset)
		if err != nil {
			return nil, err
		}
		task.Command = preset.Command
		if task.OutputFile == "" {
			task.OutputFile = preset.OutputFile
		}
		if task.Priority == 0 {
			task.Priority = preset.Priority
		}
		if preset.PreProcessing != nil && task.PreProcessing == nil {
			task.PreProcessing = &dto.NewPrePostProcessing{ScriptPath: preset.PreProcessing.ScriptPath, SidecarPath: preset.PreProcessing.SidecarPath}
		}
		if preset.PostProcessing != nil && task.PostProcessing == nil {
			task.PostProcessing = &dto.NewPrePostProcessing{ScriptPath: preset.PostProcessing.ScriptPath, SidecarPath: preset.PostProcessing.SidecarPath}
		}
	}
	t, err := s.taskRepository.Create(task, batch, source, s.sev.Session())
	if err != nil {
		return nil, err
	}

	s.sev.Metrics().Gauge("task.created").Inc()
	WebhookService().Fire(dto.TASK_CREATED, t.ToDto())
	WebsocketService().Broadcast(TASK_CREATED, t.ToDto())

	s.sev.Logger().Infof("new task added to queue (uuid: %s)", t.Uuid)
	return t, err
}

func (s *taskSvc) NewTasks(tasks *[]dto.NewTask) (*[]model.Task, error) {
	batch := uuid.NewString()
	newTasks := []model.Task{}
	for _, task := range *tasks {
		t, err := s.NewTask(&task, batch, "api")
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

	s.sev.Metrics().Gauge("batch.created").Inc()
	WebhookService().Fire(dto.BATCH_CREATED, taskDTOs)
	return &newTasks, nil
}
