package queue

import (
	"time"

	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/pkg/ffmpeg"
	"github.com/welovemedia/ffmate/pkg/service"
	"github.com/welovemedia/ffmate/sev"
	"github.com/yosev/debugo"
)

type Queue struct {
	Sev                *sev.Sev
	TaskRepository     *repository.Task
	WebhookService     *service.WebhookService
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
		q.TaskRepository.SetTaskProgress(task, progress)
	})
	if err != nil {
		q.updateTaskStatus(task, dto.DONE_ERROR)
		q.Sev.Logger().Warnf("task failed (uuid: %s):\n%v", task.Uuid, err)
		return
	}

	q.updateTaskStatus(task, dto.DONE_SUCCESSFUL)
	q.Sev.Logger().Infof("task successful (uuid: %s)", task.Uuid)
}

func (q *Queue) updateTaskStatus(task *model.Task, status dto.TaskStatus) {
	q.TaskRepository.SetTaskStatus(task, status)
	q.Sev.Metrics().Gauge("task.status.updated").Inc()
	q.WebhookService.Fire(dto.TASK_STATUS_UPDATED, task.ToDto())
}
