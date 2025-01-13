package queue

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
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

	q.updateTaskStatus(task, dto.RUNNING)
	q.Sev.Logger().Infof("processing task (uuid: %s)", task.Uuid)

	cmd := task.Command
	err := ffmpeg.Execute(&ffmpeg.ExecutionRequest{Task: task, Command: cmd, InputFile: task.InputFile, OutputFile: task.OutputFile, Logger: q.Sev.Logger()}, func(progress float64) {
		q.updateTaskProgress(task, progress)
	})
	if err != nil {
		q.updateTaskProgress(task, 100)
		q.updateTaskStatus(task, dto.DONE_ERROR)
		q.Sev.Logger().Warnf("task failed (uuid: %s):\n%v", task.Uuid, err)
		return
	}

	q.postProcessTask(task)

	q.updateTaskStatus(task, dto.DONE_SUCCESSFUL)
	q.Sev.Logger().Infof("task successful (uuid: %s)", task.Uuid)

}

func (q *Queue) postProcessTask(task *model.Task) {
	if task.PostProcessing != nil {
		q.updateTaskStatus(task, dto.POST_PROCESSING)
		if task.PostProcessing.SidecarPath != "" {
			b, err := json.Marshal(task.ToDto())
			if err != nil {
				q.Sev.Logger().Errorf("failed to marshal task: %v", err)
			} else {
				err = os.WriteFile(wildcards.Replace(task.PostProcessing.SidecarPath, task.InputFile, task.OutputFile, false), b, 0644)
				if err != nil {
					q.Sev.Logger().Errorf("failed to write task to file: %v", err)
				}
			}
		}
		args := strings.Split(wildcards.Replace(task.PostProcessing.ScriptPath, task.InputFile, task.OutputFile, true), " ")
		cmd := exec.Command(args[0], args[1:]...)
		q.Sev.Logger().Infof("triggered postProcessing (uuid: %s)", task.Uuid)

		if err := cmd.Start(); err != nil {
			q.Sev.Logger().Errorf("failed to start postProcessing (uuid: %s): %v", task.Uuid, err)
		}

		if err := cmd.Wait(); err != nil {
			q.Sev.Logger().Errorf("failed postProcessing (uuid: %s): %v", task.Uuid, err)
		}
		q.Sev.Logger().Infof("finished postProcessing (uuid: %s)", task.Uuid)
	}
}

func (q *Queue) updateTaskProgress(task *model.Task, progress float64) {
	q.TaskRepository.SetTaskProgress(task, progress)
	q.WebsocketService.Broadcast(service.TASK_UPDATED, task.ToDto())
}

func (q *Queue) updateTaskStatus(task *model.Task, status dto.TaskStatus) {
	q.TaskRepository.SetTaskStatus(task, status)
	q.WebsocketService.Broadcast(service.TASK_UPDATED, task.ToDto())
	q.Sev.Metrics().Gauge("task.status.updated").Inc()
	q.WebhookService.Fire(dto.TASK_STATUS_UPDATED, task.ToDto())
}
