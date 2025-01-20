package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/ffmpeg"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/utils/wildcards"
	"github.com/welovemedia/ffmate/sev"
	"github.com/yosev/debugo"
)

type Queue struct {
	Sev                *sev.Sev
	TaskRepository     *repository.Task
	TaskService        *service.TaskService
	WebhookService     *service.WebhookService
	WebsocketService   *service.WebsocketService
	MaxConcurrentTasks uint
}

var runningTasks = 0
var debug = debugo.New("queue")

func (q *Queue) Init() {
	go func() {
		for {
			if runningTasks < int(q.MaxConcurrentTasks) {
				task, err := q.TaskRepository.NextQueued()
				if err != nil {
					q.Sev.Logger().Errorf("failed to receive queued task from db: %v", err)
				} else if task == nil {
					debug.Debug("no queued tasks found")
				} else {

					go q.processTask(task)
				}
			} else {
				debug.Debugf("maximum concurrent tasks reached (tasks: %d/%d)", runningTasks, q.MaxConcurrentTasks)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func (q *Queue) processTask(task *model.Task) {
	runningTasks++
	defer func() { runningTasks-- }()

	task.StartedAt = time.Now().UnixMilli()
	q.Sev.Logger().Infof("processing task (uuid: %s)", task.Uuid)

	err := q.preProcessTask(task)
	if err != nil {
		q.failTask(task, fmt.Errorf("PreProcessing failed: %v", err))
		return
	}

	// resolve wildcards
	inFile := wildcards.Replace(task.InputFile.Raw, task.InputFile.Raw, task.OutputFile.Raw, false)
	outFile := wildcards.Replace(task.OutputFile.Raw, task.InputFile.Raw, task.OutputFile.Raw, false)
	task.InputFile.Resolved = inFile
	task.OutputFile.Resolved = outFile
	task.Command.Resolved = wildcards.Replace(task.Command.Raw, inFile, outFile, true)
	task.Status = dto.RUNNING
	q.updateTask(task)
	err = ffmpeg.Execute(&ffmpeg.ExecutionRequest{Task: task, Command: task.Command.Resolved, Logger: q.Sev.Logger()}, func(progress float64) {
		task.Progress = progress
		q.updateTask(task)
	})

	// task is done (successful or not)
	task.Progress = 100

	if err != nil {
		q.failTask(task, err)
		return
	}

	err = q.postProcessTask(task)
	if err != nil {
		q.failTask(task, fmt.Errorf("PostProcessing failed: %v", err))
		return
	}

	task.FinishedAt = time.Now().UnixMilli()
	task.Status = dto.DONE_SUCCESSFUL
	q.updateTask(task)
	q.Sev.Logger().Infof("task successful (uuid: %s)", task.Uuid)

}

func (q *Queue) preProcessTask(task *model.Task) error {
	if task.PreProcessing != nil && (task.PreProcessing.SidecarPath != nil || task.PreProcessing.ScriptPath != nil) {
		q.Sev.Logger().Infof("starting preProcessing (uuid: %s)", task.Uuid)
		task.PreProcessing.StartedAt = time.Now().UnixMilli()
		task.Status = dto.PRE_PROCESSING
		q.updateTask(task)
		if task.PreProcessing.SidecarPath != nil && task.PreProcessing.SidecarPath.Raw != "" {
			b, err := json.Marshal(task.ToDto())
			if err != nil {
				q.Sev.Logger().Errorf("failed to marshal task to write sidecar file: %v", err)
			} else {
				task.PreProcessing.SidecarPath.Resolved = wildcards.Replace(task.PreProcessing.SidecarPath.Raw, task.InputFile.Raw, task.OutputFile.Raw, false)
				q.updateTask(task)
				err = os.WriteFile(task.PreProcessing.SidecarPath.Resolved, b, 0644)
				if err != nil {
					task.PreProcessing.Error = fmt.Errorf("failed to write sidecar: %v", err).Error()
					q.Sev.Logger().Errorf("failed to write sidecar file: %v", err)
				} else {
					debug.Debugf("wrote sidecar file (uuid: %s)", task.Uuid)
				}
			}
		}

		if task.PreProcessing.Error == "" && task.PreProcessing.ScriptPath != nil && task.PreProcessing.ScriptPath.Raw != "" {
			task.PreProcessing.ScriptPath.Resolved = wildcards.Replace(task.PreProcessing.ScriptPath.Raw, task.InputFile.Raw, task.OutputFile.Raw, true)
			q.updateTask(task)
			args := ffmpeg.SplitCommand(task.PreProcessing.ScriptPath.Resolved)
			cmd := exec.Command(args[0], args[1:]...)
			debug.Debugf("triggered preProcessing script (uuid: %s)", task.Uuid)

			if err := cmd.Start(); err != nil {
				task.PreProcessing.Error = err.Error()
				q.Sev.Logger().Errorf("failed to start preProcessing script (uuid: %s): %v", task.Uuid, err)
			} else {
				if err := cmd.Wait(); err != nil {
					task.PreProcessing.Error = err.Error()
					q.Sev.Logger().Errorf("failed preProcessing script (uuid: %s): %v", task.Uuid, err)
				}

			}
		}

		task.PreProcessing.FinishedAt = time.Now().UnixMilli()
		if task.PreProcessing.Error != "" {
			q.Sev.Logger().Infof("finished preProcessing with error (uuid: %s)", task.Uuid)
			return errors.New(task.PreProcessing.Error)
		}
		q.Sev.Logger().Infof("finished preProcessing (uuid: %s)", task.Uuid)
	}
	return nil
}

func (q *Queue) postProcessTask(task *model.Task) error {
	if task.PostProcessing != nil && (task.PostProcessing.SidecarPath != nil || task.PostProcessing.ScriptPath != nil) {
		q.Sev.Logger().Infof("starting postProcessing (uuid: %s)", task.Uuid)
		task.PostProcessing.StartedAt = time.Now().UnixMilli()
		task.Status = dto.POST_PROCESSING
		q.updateTask(task)
		if task.PostProcessing.SidecarPath != nil && task.PostProcessing.SidecarPath.Raw != "" {
			b, err := json.Marshal(task.ToDto())
			if err != nil {
				q.Sev.Logger().Errorf("failed to marshal task to write sidecar file: %v", err)
			} else {
				task.PostProcessing.SidecarPath.Resolved = wildcards.Replace(task.PostProcessing.SidecarPath.Raw, task.InputFile.Resolved, task.OutputFile.Resolved, false)
				q.updateTask(task)
				err = os.WriteFile(task.PostProcessing.SidecarPath.Resolved, b, 0644)
				if err != nil {
					task.PostProcessing.Error = fmt.Errorf("failed to write sidecar: %v", err).Error()
					q.Sev.Logger().Errorf("failed to write sidecar file: %v", err)
				} else {
					debug.Debugf("wrote sidecar file (uuid: %s)", task.Uuid)
				}
			}
		}

		if task.PostProcessing.Error == "" && task.PostProcessing.ScriptPath != nil && task.PostProcessing.ScriptPath.Raw != "" {
			task.PostProcessing.ScriptPath.Resolved = wildcards.Replace(task.PostProcessing.ScriptPath.Raw, task.InputFile.Resolved, task.OutputFile.Resolved, true)
			q.updateTask(task)
			args := ffmpeg.SplitCommand(task.PostProcessing.ScriptPath.Resolved)
			cmd := exec.Command(args[0], args[1:]...)
			debug.Debugf("triggered postProcessing script (uuid: %s)", task.Uuid)

			if err := cmd.Start(); err != nil {
				task.PostProcessing.Error = err.Error()
				q.Sev.Logger().Errorf("failed to start postProcessing script (uuid: %s): %v", task.Uuid, err)
			} else {
				if err := cmd.Wait(); err != nil {
					task.PostProcessing.Error = err.Error()
					q.Sev.Logger().Errorf("failed postProcessing script (uuid: %s): %v", task.Uuid, err)
				}

			}
		}

		task.PostProcessing.FinishedAt = time.Now().UnixMilli()
		if task.PostProcessing.Error != "" {
			q.Sev.Logger().Infof("finished postProcessing with error (uuid: %s)", task.Uuid)
			return errors.New(task.PostProcessing.Error)
		}
		q.Sev.Logger().Infof("finished postProcessing (uuid: %s)", task.Uuid)
		return nil
	}
	return nil
}

func (q *Queue) failTask(task *model.Task, err error) {
	task.FinishedAt = time.Now().UnixMilli()
	task.Progress = 100
	task.Status = dto.DONE_ERROR
	task.Error = err.Error()
	q.updateTask(task)
	q.Sev.Logger().Warnf("task failed (uuid: %s):\n%v", task.Uuid, err)
}

func (q *Queue) updateTask(task *model.Task) {
	q.TaskService.UpdateTask(task)
}
