package task

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-shellwords"
	"github.com/tidwall/gjson"
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
	CountAllStatus(session string) (int, int, int, int, int, error)
	NextQueued(int) (*[]model.Task, error)
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

	s.webhookService.Fire(dto.TASK_UPDATED, task.ToDto())
	s.webhookService.FireDirect(task.Webhooks, dto.TASK_UPDATED, task.ToDto())
	s.websocketService.Broadcast(websocket.TASK_UPDATED, task.ToDto())

	if task.Batch != "" {
		switch task.Status {
		case dto.DONE_SUCCESSFUL, dto.DONE_ERROR, dto.DONE_CANCELED:
			c, _ := s.repository.CountUnfinishedByBatch(task.Batch)
			if c == 0 {
				metrics.Gauge("batch.finished").Inc()
				s.webhookService.Fire(dto.BATCH_FINISHED, task.ToDto())
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

	w.Status = dto.DONE_CANCELED
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

	w.Status = dto.QUEUED
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
		taskDTOs = append(taskDTOs, task.ToDto())
	}

	return &dto.Batch{
		Uuid:  uuid,
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

	task := &model.Task{
		Uuid:             uuid.NewString(),
		Command:          &dto.RawResolved{Raw: newTask.Command},
		InputFile:        &dto.RawResolved{Raw: newTask.InputFile},
		OutputFile:       &dto.RawResolved{Raw: newTask.OutputFile},
		Metadata:         newTask.Metadata,
		Name:             newTask.Name,
		Priority:         newTask.Priority,
		Progress:         0,
		Source:           source,
		Status:           dto.QUEUED,
		Batch:            batch,
		Webhooks:         newTask.Webhooks,
		ClientIdentifier: cfg.GetString("ffmate.identifier"),
	}
	w, err := s.repository.Add(task)
	debug.Task.Info("created task (uuid: %s)", w.Uuid)

	metrics.Gauge("task.created").Inc()
	s.webhookService.Fire(dto.TASK_CREATED, w.ToDto())
	s.webhookService.FireDirect(w.Webhooks, dto.TASK_CREATED, w.ToDto())
	s.websocketService.Broadcast(websocket.TASK_CREATED, w.ToDto())

	return w, err
}

func (s *Service) AddBatch(newBatch *dto.NewBatch, source dto.TaskSource) (*dto.Batch, error) {
	batchUuid := uuid.NewString()
	tasks := []model.Task{}
	for _, task := range newBatch.Tasks {
		t, err := s.Add(task, "api", batchUuid)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}

	// clear preset cache
	presetCache.Delete(batchUuid)

	// transform each task to its DTO
	var taskDTOs = []*dto.Task{}
	for _, task := range tasks {
		taskDTOs = append(taskDTOs, task.ToDto())
	}

	metrics.Gauge("batch.created").Inc()
	s.webhookService.Fire(dto.BATCH_CREATED, taskDTOs)

	batch := &dto.Batch{
		Uuid:  batchUuid,
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
	s.webhookService.Fire(dto.TASK_DELETED, w.ToDto())
	s.webhookService.FireDirect(w.Webhooks, dto.TASK_DELETED, w.ToDto())
	s.websocketService.Broadcast(websocket.TASK_DELETED, w.ToDto())

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
		var maxConcurrentTasks int
		maxConcurrentTasks = cfg.GetInt("ffmate.maxConcurrentTasks")
		if maxConcurrentTasks == 0 || maxConcurrentTasks <= taskQueueLength {
			debug.Task.Debug("maximum concurrent tasks reached (tasks: %d/%d)", taskQueueLength, maxConcurrentTasks)
			continue
		}

		task, err := s.repository.NextQueued(maxConcurrentTasks - taskQueueLength)
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
			taskQueue.Store(t.Uuid, ctx)
			go s.processNewTask(&t)
		}
	}
}

func (s *Service) processNewTask(task *model.Task) {
	debug.Task.Info("processing task (uuid: %s)", task.Uuid)
	defer taskQueue.Delete(task.Uuid)

	task.StartedAt = time.Now().UnixMilli()

	// preProcessing
	err := s.prePostProcessTask(task, task.PreProcessing, "pre")
	if err != nil {
		s.failTask(task, fmt.Errorf("PreProcessing failed: %v", err))
		return
	}

	// resolve wildcards
	inFile := s.wildcardReplacer(task.InputFile.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
	outFile := s.wildcardReplacer(task.OutputFile.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
	task.InputFile.Resolved = inFile
	task.OutputFile.Resolved = outFile
	task.Command.Resolved = s.wildcardReplacer(task.Command.Raw, inFile, outFile, task.Source, task.Metadata)
	task.Status = dto.RUNNING
	s.Update(task)

	// create output directory if it does not exist (recursive)
	err = os.MkdirAll(filepath.Dir(task.OutputFile.Resolved), 0755)
	if err != nil {
		s.failTask(task, fmt.Errorf("failed to create non-existing output directory: %v", err))
		return
	}

	// process with ffmpeg
	debug.Task.Debug("starting ffmpeg process (uuid: %s)", task.Uuid)
	ctxAny, _ := taskQueue.Load(task.Uuid)
	ctx := ctxAny.(context.Context)
	err = s.ffmpegService.Execute(
		&ffmpeg.ExecutionRequest{
			Task:    task,
			Command: task.Command.Resolved,
			Ctx:     ctx,
			UpdateFunc: func(progress float64, remaining float64) {
				task.Progress = progress
				task.Remaining = remaining
				s.Update(task)
			},
		},
	)

	// task is done (successful or not)
	task.Progress = 100
	task.Remaining = -1
	if err != nil {
		debug.Task.Debug("finished processing with error (uuid: %s): %v", task.Uuid, err)
		if context.Cause(ctx) != nil {
			s.cancelTask(task, context.Cause(ctx))
			return
		}
		s.failTask(task, err)
		return
	}

	debug.Task.Debug("finished processing (uuid: %s)", task.Uuid)

	err = s.prePostProcessTask(task, task.PostProcessing, "post")
	if err != nil {
		s.failTask(task, fmt.Errorf("PostProcessing failed: %v", err))
		return
	}

	task.FinishedAt = time.Now().UnixMilli()
	task.Status = dto.DONE_SUCCESSFUL
	s.Update(task)
	debug.Task.Info("task successful (uuid: %s)", task.Uuid)
}

func (s *Service) prePostProcessTask(task *model.Task, processor *dto.PrePostProcessing, processorType string) error {
	if processor != nil && (processor.SidecarPath != nil || processor.ScriptPath != nil) {
		if processorType == "pre" {
			metrics.GaugeVec("task.preProcessing").WithLabelValues(strconv.FormatBool(processor.SidecarPath != nil && processor.SidecarPath.Raw == ""), strconv.FormatBool(processor.ScriptPath != nil && processor.ScriptPath.Raw == "")).Inc()
		} else {
			metrics.GaugeVec("task.postProcessing").WithLabelValues(strconv.FormatBool(processor.SidecarPath != nil && processor.SidecarPath.Raw == ""), strconv.FormatBool(processor.ScriptPath != nil && processor.ScriptPath.Raw == "")).Inc()
		}
		debug.Task.Debug("starting %sProcessing (uuid: %s)", processorType, task.Uuid)
		processor.StartedAt = time.Now().UnixMilli()
		if processorType == "pre" {
			task.Status = dto.PRE_PROCESSING
		} else {
			task.Status = dto.POST_PROCESSING
		}
		s.Update(task)
		if processor.SidecarPath != nil && processor.SidecarPath.Raw != "" {
			b, err := json.Marshal(task.ToDto())
			if err != nil {
				debug.Log.Error("failed to marshal task to write sidecar file: %v", err)
			} else {
				if processorType == "pre" {
					processor.SidecarPath.Resolved = s.wildcardReplacer(processor.SidecarPath.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
				} else {
					processor.SidecarPath.Resolved = s.wildcardReplacer(processor.SidecarPath.Raw, task.InputFile.Resolved, task.OutputFile.Resolved, task.Source, task.Metadata)
				}
				s.Update(task)

				// create sidebar-output directory if it does not exist (recursive)
				err := os.MkdirAll(filepath.Dir(processor.SidecarPath.Resolved), 0755)
				if err != nil {
					return err
				}

				err = os.WriteFile(processor.SidecarPath.Resolved, b, 0644)
				if err != nil {
					processor.Error = fmt.Errorf("failed to write sidecar: %v", err).Error()
					debug.Log.Error("failed to write sidecar file: %v", err)
				} else {
					debug.Task.Debug("wrote sidecar file (uuid: %s)", task.Uuid)
				}
			}
		}

		if processor.Error == "" && processor.ScriptPath != nil && processor.ScriptPath.Raw != "" {
			if processorType == "pre" {
				processor.ScriptPath.Resolved = s.wildcardReplacer(processor.ScriptPath.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
			} else {
				processor.ScriptPath.Resolved = s.wildcardReplacer(processor.ScriptPath.Raw, task.InputFile.Resolved, task.OutputFile.Resolved, task.Source, task.Metadata)
			}
			s.Update(task)
			args, err := shellwords.NewParser().Parse(processor.ScriptPath.Resolved)
			if err != nil {
				processor.Error = err.Error()
				debug.Task.Debug("failed to parse %sProcessing script (uuid: %s): %v", processorType, task.Uuid, err)
			} else {
				cmd := exec.Command(args[0], args[1:]...)
				debug.Task.Debug("triggered %sProcessing script (uuid: %s)", processorType, task.Uuid)

				var stderr bytes.Buffer
				cmd.Stderr = &stderr

				if err := cmd.Start(); err != nil {
					processor.Error = fmt.Sprintf("%s (exit code: %d)", stderr.String(), cmd.ProcessState.ExitCode())
					debug.Task.Debug("failed to start %sProcessing script with exit code %d (uuid: %s): stderr: %s", processorType, cmd.ProcessState.ExitCode(), task.Uuid, stderr.String())
				} else {
					if err := cmd.Wait(); err != nil {
						processor.Error = fmt.Sprintf("%s (exit code: %d)", stderr.String(), cmd.ProcessState.ExitCode())
						debug.Task.Debug("failed %sProcessing script with exit code %d (uuid: %s): stderr: %s", processorType, cmd.ProcessState.ExitCode(), task.Uuid, stderr.String())
					}
				}
			}
		}

		// re-import the sidecar file and unmarshal into task
		// enabled modifying the task from within a preProcess script by modifying the sideCar file before re-importing it
		if processorType == "pre" && processor.SidecarPath != nil && processor.SidecarPath.Raw != "" && processor.ImportSidecar {
			b, err := os.ReadFile(processor.SidecarPath.Resolved)
			if err != nil {
				return err
			}
			err = json.Unmarshal(b, task)
			if err != nil {
				return err
			}
			debug.Task.Debug("re-imported sidecar file (uuid: %s)", task.Uuid)
		}

		processor.FinishedAt = time.Now().UnixMilli()
		if processor.Error != "" {
			debug.Task.Info("finished %sProcessing with error (uuid: %s)", processorType, task.Uuid)
			return errors.New(processor.Error)
		}
		debug.Task.Info("finished %sProcessing (uuid: %s)", processorType, task.Uuid)
	}
	return nil
}

func (s *Service) wildcardReplacer(input string, inputFile string, outputFile string, source dto.TaskSource, metadata *dto.MetadataMap) string {
	input = strings.ReplaceAll(input, "${INPUT_FILE}", fmt.Sprintf("\"%s\"", inputFile))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE}", fmt.Sprintf("\"%s\"", outputFile))

	input = strings.ReplaceAll(input, "${INPUT_FILE_BASE}", filepath.Base(inputFile))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_BASE}", filepath.Base(inputFile))
	input = strings.ReplaceAll(input, "${INPUT_FILE_EXTENSION}", filepath.Ext(filepath.Base(inputFile)))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_EXTENSION", filepath.Ext(filepath.Base(outputFile)))
	input = strings.ReplaceAll(input, "${INPUT_FILE_BASENAME}", strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(filepath.Base(inputFile))))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_BASENAME}", strings.TrimSuffix(filepath.Base(outputFile), filepath.Ext(filepath.Base(outputFile))))
	input = strings.ReplaceAll(input, "${INPUT_FILE_DIR}", filepath.Dir(inputFile))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_DIR}", filepath.Dir(inputFile))

	input = strings.ReplaceAll(input, "${DATE_YEAR}", time.Now().Format("2006"))
	input = strings.ReplaceAll(input, "${DATE_SHORTYEAR}", time.Now().Format("06"))
	input = strings.ReplaceAll(input, "${DATE_MONTH}", time.Now().Format("01"))
	input = strings.ReplaceAll(input, "${DATE_DAY}", time.Now().Format("02"))

	_, week := time.Now().ISOWeek()
	input = strings.ReplaceAll(input, "${DATE_WEEK}", strconv.Itoa(week))

	input = strings.ReplaceAll(input, "${TIME_HOUR}", time.Now().Format("15"))
	input = strings.ReplaceAll(input, "${TIME_MINUTE}", time.Now().Format("04"))
	input = strings.ReplaceAll(input, "${TIME_SECOND}", time.Now().Format("05"))

	input = strings.ReplaceAll(input, "${TIMESTAMP_SECONDS}", strconv.FormatInt(time.Now().Unix(), 10))
	input = strings.ReplaceAll(input, "${TIMESTAMP_MILLISECONDS}", strconv.FormatInt(time.Now().UnixMilli(), 10))
	input = strings.ReplaceAll(input, "${TIMESTAMP_MICROSECONDS}", strconv.FormatInt(time.Now().UnixMicro(), 10))
	input = strings.ReplaceAll(input, "${TIMESTAMP_NANOSECONDS}", strconv.FormatInt(time.Now().UnixNano(), 10))

	input = strings.ReplaceAll(input, "${OS_NAME}", runtime.GOOS)
	input = strings.ReplaceAll(input, "${OS_ARCH}", runtime.GOARCH)

	input = strings.ReplaceAll(input, "${SOURCE}", string(source))

	input = strings.ReplaceAll(input, "${UUID}", uuid.NewString())

	input = strings.ReplaceAll(input, "${FFMPEG}", cfg.GetString("ffmate.ffmpeg"))

	// handle metadata wildcard
	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)

		if err == nil {
			jsonStr := string(metadataJSON)
			re := regexp.MustCompile(`\$\{METADATA_([^}]+)\}`)
			input = re.ReplaceAllStringFunc(input, func(match string) string {
				path := re.FindStringSubmatch(match)[1]
				val := gjson.Get(jsonStr, path)
				if val.Exists() {
					return val.String()
				}
				return ""
			})
		}
	}

	return input
}

func (s *Service) cancelTask(task *model.Task, err error) {
	task.FinishedAt = time.Now().UnixMilli()
	task.Progress = 100
	task.Status = dto.DONE_CANCELED
	task.Error = err.Error()
	s.Update(task)
	debug.Task.Info("task canceled (uuid: %s): %v", task.Uuid, err)
}

func (s *Service) failTask(task *model.Task, err error) {
	task.FinishedAt = time.Now().UnixMilli()
	task.Progress = 100
	task.Status = dto.DONE_ERROR
	task.Error = err.Error()
	s.Update(task)
	debug.Task.Warn("task failed (uuid: %s)", task.Uuid)
}

func (s *Service) taskQueueLength() int {
	length := 0
	taskQueue.Range(func(key, value any) bool {
		length++
		return true
	})
	return length
}

/**
 * CountAllStatus is used in systray
 */

func (s *Service) CountAllStatus(session bool) (queued, running, doneSuccessful, doneError, doneCanceled int, err error) {
	if session {
		return s.repository.CountAllStatus(cfg.GetString("ffmate.session"))
	}
	return s.repository.CountAllStatus("")
}

func (s *Service) Name() string {
	return service.Task
}
