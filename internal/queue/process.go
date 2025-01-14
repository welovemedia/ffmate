package queue

import (
	"encoding/json"
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

	task.Status = dto.RUNNING
	task.StartedAt = time.Now().UnixMilli()
	q.updateTask(task)
	q.Sev.Logger().Infof("processing task (uuid: %s)", task.Uuid)

	// resolve wildcards
	inFile := wildcards.Replace(task.InputFile, task.InputFile, task.OutputFile, false)
	outFile := wildcards.Replace(task.OutputFile, task.InputFile, task.OutputFile, false)
	task.Resolved = &dto.Resolved{
		InputFile:  inFile,
		OutputFile: outFile,
		Command:    wildcards.Replace(task.Command, inFile, outFile, true),
	}

	err := ffmpeg.Execute(&ffmpeg.ExecutionRequest{Task: task, Command: task.Resolved.Command, Logger: q.Sev.Logger()}, func(progress float64) {
		task.Progress = progress
		q.updateTask(task)
	})

	// task is done (successful or not)
	task.Progress = 100

	if err != nil {
		task.FinishedAt = time.Now().UnixMilli()
		task.Status = dto.DONE_ERROR
		task.Error = err.Error()
		q.updateTask(task)
		q.Sev.Logger().Warnf("task failed (uuid: %s):\n%v", task.Uuid, err)
		return
	}

	q.postProcessTask(task)

	task.FinishedAt = time.Now().UnixMilli()
	task.Status = dto.DONE_SUCCESSFUL
	q.updateTask(task)
	q.Sev.Logger().Infof("task successful (uuid: %s)", task.Uuid)

}

func (q *Queue) postProcessTask(task *model.Task) {
	if task.PostProcessing != nil {
		q.Sev.Logger().Infof("starting postProcessing (uuid: %s)", task.Uuid)
		task.PostProcessing.StartedAt = time.Now().UnixMilli()
		task.Status = dto.POST_PROCESSING
		q.updateTask(task)
		if task.PostProcessing.SidecarPath != "" {
			b, err := json.Marshal(task.ToDto())
			if err != nil {
				q.Sev.Logger().Errorf("failed to marshal task to write sidecar file: %v", err)
			} else {
				task.Resolved.PostProcessing.SidecarPath = wildcards.Replace(task.PostProcessing.SidecarPath, task.Resolved.InputFile, task.Resolved.OutputFile, false)
				err = os.WriteFile(task.Resolved.PostProcessing.ScriptPath, b, 0644)
				if err != nil {
					q.Sev.Logger().Errorf("failed to write sidecar file: %v", err)
				} else {
					debug.Debug("wrote sidecar file (uuid: %s)", task.Uuid)
				}

			}
		}

		if task.PostProcessing.ScriptPath != "" {
			task.Resolved.PostProcessing.ScriptPath = wildcards.Replace(task.PostProcessing.ScriptPath, task.Resolved.InputFile, task.Resolved.OutputFile, true)
			args := ffmpeg.SplitCommand(task.Resolved.PostProcessing.ScriptPath)
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
		q.Sev.Logger().Infof("finished postProcessing (uuid: %s)", task.Uuid)
	}
}

func (q *Queue) updateTask(task *model.Task) {
	q.TaskRepository.UpdateTask(task)
	q.WebsocketService.Broadcast(service.TASK_UPDATED, task.ToDto())
	q.Sev.Metrics().Gauge("task.status.updated").Inc()
	q.WebhookService.Fire(dto.TASK_STATUS_UPDATED, task.ToDto())
}
