package task

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/ffmpeg"
	"github.com/welovemedia/ffmate/v2/internal/service/preset"
	"github.com/welovemedia/ffmate/v2/internal/service/webhook"
	"github.com/welovemedia/ffmate/v2/internal/service/websocket"
)

type Repository interface {
	List(page int, perPage int) (*[]model.Task, int64, error)
	ListByBatch(uuid string, page int, perPage int) (*[]model.Task, int64, error)
	Add(task *model.Task) (*model.Task, error)
	Update(task *model.Task) (*model.Task, error)
	First(uuid string) (*model.Task, error)
	Delete(task *model.Task) error
	Count() (int64, error)
	CountUnfinishedByBatch(uuid string) (int64, error)
	CountAllStatus() (int, int, int, int, int, error)
	NextQueued(amount int, labels dto.Labels) (*[]model.Task, error)
}

type Service struct {
	repository       Repository
	presetService    *preset.Service
	webhookService   *webhook.Service
	websocketService *websocket.Service
	ffmpegService    *ffmpeg.Service
}

func NewService(repository Repository, presetService *preset.Service, webhookService *webhook.Service, websocketService *websocket.Service, ffmpegService *ffmpeg.Service) *Service {
	return &Service{
		repository:       repository,
		presetService:    presetService,
		webhookService:   webhookService,
		websocketService: websocketService,
		ffmpegService:    ffmpegService,
	}
}

func (s *Service) Get(uuid string) (*model.Task, error) {
	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("task for given uuid not found")
	}

	return w, nil
}

func (s *Service) Update(task *model.Task) (*model.Task, error) {
	task.ClientIdentifier = cfg.GetString("ffmate.identifier")
	task, err := s.repository.Update(task)
	if err != nil {
		return nil, err
	}

	s.webhookService.Fire(dto.TaskUpdated, task.ToDTO())
	s.webhookService.FireDirect(task.Webhooks, dto.TaskUpdated, task.ToDTO())
	s.websocketService.Broadcast(websocket.TaskUpdated, task.ToDTO())

	if task.Batch != "" {
		switch task.Status {
		case dto.DoneSuccessful, dto.DoneError, dto.DoneCanceled:
			c, _ := s.repository.CountUnfinishedByBatch(task.Batch)
			if c == 0 {
				metrics.Gauge("batch.finished").Inc()
				s.webhookService.Fire(dto.BatchFinished, task.ToDTO())
			}
		}
	}

	return task, nil
}

func (s *Service) Cancel(uuid string) (*model.Task, error) {
	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("task for given uuid not found")
	}

	w.Status = dto.DoneCanceled
	w.Remaining = -1
	w.Progress = 100
	w.FinishedAt = time.Now().UnixMilli()

	metrics.Gauge("task.canceled").Inc()
	debug.Log.Info("canceled task (uuid: %s)", uuid)

	return s.Update(w)
}

func (s *Service) Restart(uuid string) (*model.Task, error) {
	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("task for given uuid not found")
	}

	w.Status = dto.Queued
	w.Progress = 0
	w.StartedAt = 0
	w.FinishedAt = 0
	w.Error = ""

	metrics.Gauge("task.restarted").Inc()
	debug.Log.Info("restarted task (uuid: %s)", uuid)

	return s.Update(w)
}

func (s *Service) List(page int, perPage int) (*[]model.Task, int64, error) {
	return s.repository.List(page, perPage)
}

func (s *Service) GetBatch(uuid string, page int, perPage int) (*dto.Batch, int64, error) {
	tasks, count, err := s.repository.ListByBatch(uuid, page, perPage)
	if err != nil {
		return nil, count, err
	}

	var taskDTOs = []*dto.Task{}
	for _, task := range *tasks {
		taskDTOs = append(taskDTOs, task.ToDTO())
	}

	return &dto.Batch{
		UUID:  uuid,
		Tasks: taskDTOs,
	}, count, err
}

var presetCache = sync.Map{}

func (s *Service) Add(newTask *dto.NewTask, source dto.TaskSource, batch string) (*model.Task, error) {
	if newTask.Preset != "" {
		var preset *model.Preset
		var err error

		// add preset cache for batch creation
		if batch != "" {
			p, ok := presetCache.Load(batch)
			if ok {
				preset = p.(*model.Preset)
			}
		}
		if preset == nil {
			preset, err = s.presetService.Get(newTask.Preset)
			if err != nil {
				return nil, err
			}

			if preset != nil {
				presetCache.Store(batch, preset)
			}
		}

		newTask.Command = preset.Command
		if newTask.OutputFile == "" {
			newTask.OutputFile = preset.OutputFile
		}
		if newTask.Priority == 0 {
			newTask.Priority = preset.Priority
		}
		if preset.PreProcessing != nil && newTask.PreProcessing == nil {
			newTask.PreProcessing = &dto.NewPrePostProcessing{ScriptPath: preset.PreProcessing.ScriptPath, SidecarPath: preset.PreProcessing.SidecarPath, ImportSidecar: preset.PreProcessing.ImportSidecar}
		}
		if preset.PostProcessing != nil && newTask.PostProcessing == nil {
			newTask.PostProcessing = &dto.NewPrePostProcessing{ScriptPath: preset.PostProcessing.ScriptPath, SidecarPath: preset.PostProcessing.SidecarPath}
		}

		// apply labels from preset if no direct labels are set
		if len(newTask.Labels) == 0 {
			var labels = make(dto.Labels, len(preset.Labels))
			for i, label := range preset.Labels {
				labels[i] = label.Value
			}
			newTask.Labels = labels
		}

		if preset.Webhooks != nil {
			if newTask.Webhooks == nil {
				newTask.Webhooks = preset.Webhooks
			} else {
				*newTask.Webhooks = append(*preset.Webhooks, *newTask.Webhooks...)
			}
		}
	}

	// filter webhooks so only "task.*" events remain
	if newTask.Webhooks != nil {
		filtered := make(dto.DirectWebhooks, 0, len(*newTask.Webhooks))
		for _, wh := range *newTask.Webhooks {
			if strings.HasPrefix(string(wh.Event), "task.") {
				filtered = append(filtered, wh)
			}
		}
		newTask.Webhooks = &filtered
	}

	var labels = make([]model.Label, len(newTask.Labels))
	for i, label := range newTask.Labels {
		labels[i] = model.Label{Value: label}
	}

	task := &model.Task{
		UUID:             uuid.NewString(),
		Command:          &dto.RawResolved{Raw: newTask.Command},
		InputFile:        &dto.RawResolved{Raw: newTask.InputFile},
		OutputFile:       &dto.RawResolved{Raw: newTask.OutputFile},
		Metadata:         newTask.Metadata,
		Name:             newTask.Name,
		Priority:         newTask.Priority,
		Progress:         0,
		Source:           source,
		Labels:           labels,
		Status:           dto.Queued,
		Batch:            batch,
		Webhooks:         newTask.Webhooks,
		ClientIdentifier: cfg.GetString("ffmate.identifier"),
	}
	w, err := s.repository.Add(task)
	debug.Task.Info("created task (uuid: %s)", w.UUID)

	metrics.Gauge("task.created").Inc()
	s.webhookService.Fire(dto.TaskCreated, w.ToDTO())
	s.webhookService.FireDirect(w.Webhooks, dto.TaskCreated, w.ToDTO())
	s.websocketService.Broadcast(websocket.TaskCreated, w.ToDTO())

	return w, err
}

func (s *Service) AddBatch(newBatch *dto.NewBatch) (*dto.Batch, error) {
	batchUUID := uuid.NewString()
	tasks := []model.Task{}
	for _, task := range newBatch.Tasks {
		t, err := s.Add(task, "api", batchUUID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}

	// clear preset cache
	presetCache.Delete(batchUUID)

	// transform each task to its DTO
	var taskDTOs = []*dto.Task{}
	for _, task := range tasks {
		taskDTOs = append(taskDTOs, task.ToDTO())
	}

	metrics.Gauge("batch.created").Inc()
	s.webhookService.Fire(dto.BatCreated, taskDTOs)

	batch := &dto.Batch{
		UUID:  batchUUID,
		Tasks: taskDTOs,
	}

	return batch, nil
}

func (s *Service) Delete(uuid string) error {
	w, err := s.repository.First(uuid)
	if err != nil {
		return err
	}

	if w == nil {
		return errors.New("task for given uuid not found")
	}

	err = s.repository.Delete(w)
	if err != nil {
		debug.Log.Error("failed to delete task (uuid: %s)", uuid)
		return err
	}

	debug.Log.Info("deleted task (uuid: %s)", uuid)

	metrics.Gauge("task.deleted").Inc()
	s.webhookService.Fire(dto.TaskDeleted, w.ToDTO())
	s.webhookService.FireDirect(w.Webhooks, dto.TaskDeleted, w.ToDTO())
	s.websocketService.Broadcast(websocket.TaskDeleted, w.ToDTO())

	return nil
}

/**
 * Task processing
 */

var taskQueue = sync.Map{}

func (s *Service) ProcessQueue() *Service {
	// lookup ffmpeg (path)
	if !s.checkFFmpeg() {
		go func() {
			for {
				time.Sleep(10 * time.Second)
				if s.checkFFmpeg() {
					return
				}
			}
		}()
	}

	go s.processQueue()

	return s
}

func (s *Service) checkFFmpeg() bool {
	if !cfg.Has("ffmate.ffmpeg") || cfg.GetString("ffmate.ffmpeg") == "" {
		cfg.Set("ffmate.ffmpeg", "ffmpeg")
	}
	if path, err := exec.LookPath(cfg.GetString("ffmate.ffmpeg")); err != nil {
		debug.Task.Error("ffmpeg binary not found in PATH. Please install ffmpeg or set the path to the ffmpeg binary with the --ffmpeg flag. Error: %s", err)
	} else {
		cfg.Set("ffmate.ffmpeg", path)
		cfg.Set("ffmate.isFFmpeg", true)
		debug.Log.Info("ffmpeg binary found at '%s'", cfg.GetString("ffmate.ffmpeg"))
		return true
	}
	return false
}

func (s *Service) processQueue() {
	for {
		time.Sleep(1 * time.Second)

		if !cfg.GetBool("ffmate.isFFmpeg") {
			debug.Task.Debug("ffmpeg not configured yet, skipping processing")
			continue
		}

		taskQueueLength := s.taskQueueLength()
		var maxConcurrentTasks = cfg.GetInt("ffmate.maxConcurrentTasks")
		if maxConcurrentTasks == 0 || maxConcurrentTasks <= taskQueueLength {
			debug.Task.Debug("maximum concurrent tasks reached (tasks: %d/%d)", taskQueueLength, maxConcurrentTasks)
			continue
		}

		task, err := s.repository.NextQueued(maxConcurrentTasks-taskQueueLength, cfg.GetStringSlice("ffmate.labels"))
		if err != nil {
			debug.Log.Error("failed to receive queued task from db: %v", err)
			continue
		}
		if task == nil || len(*task) == 0 {
			debug.Task.Debug("no queued tasks found")
			continue
		}

		for _, t := range *task {
			ctx := context.Background()
			taskQueue.Store(t.UUID, ctx)
			go s.processNewTask(&t)
		}
	}
}

func (s *Service) processNewTask(task *model.Task) {
	debug.Task.Info("processing task (uuid: %s)", task.UUID)
	defer taskQueue.Delete(task.UUID)

	task.StartedAt = time.Now().UnixMilli()

	if err := s.runPreProcessing(task); err != nil {
		s.failTask(task, err)
		return
	}

	s.prepareTaskFiles(task)

	if err := s.createOutputDirectory(task); err != nil {
		s.failTask(task, err)
		return
	}

	if err := s.executeFFmpeg(task); err != nil {
		return // failure already handled inside executeFFmpeg
	}

	if err := s.runPostProcessing(task); err != nil {
		s.failTask(task, err)
		return
	}

	s.finalizeTask(task)
}

func (s *Service) cancelTask(task *model.Task, err error) {
	task.FinishedAt = time.Now().UnixMilli()
	task.Progress = 100
	task.Status = dto.DoneCanceled
	task.Error = err.Error()
	_, err = s.Update(task)
	if err != nil {
		debug.Task.Error("failed to update task after cancel (uuid: %s)", task.UUID)
	}
	debug.Task.Info("task canceled (uuid: %s): %v", task.UUID, err)
}

func (s *Service) failTask(task *model.Task, err error) {
	task.FinishedAt = time.Now().UnixMilli()
	task.Progress = 100
	task.Status = dto.DoneError
	task.Error = err.Error()
	_, err = s.Update(task)
	if err != nil {
		debug.Task.Error("failed to update task after fail (uuid: %s)", task.UUID)
	}
	debug.Task.Warn("task failed (uuid: %s)", task.UUID)
}

func (s *Service) taskQueueLength() int {
	length := 0
	taskQueue.Range(func(_, _ any) bool {
		length++
		return true
	})
	return length
}

/**
 * CountAllStatus is used in systray
 */

func (s *Service) CountAllStatus() (queued, running, doneSuccessful, doneError, doneCanceled int, err error) {
	return s.repository.CountAllStatus()
}

func (s *Service) Name() string {
	return service.Task
}
