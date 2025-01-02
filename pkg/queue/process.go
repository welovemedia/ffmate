package queue

import (
	"time"

	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/pkg/ffmpeg"
	"github.com/welovemedia/ffmate/pkg/service"
	"github.com/welovemedia/ffmate/sev"
)

type Queue struct {
	Sev                *sev.Sev
	TaskRepository     *repository.Task
	WebhookService     *service.WebhookService
	MaxConcurrentTasks uint
}

var runningTasks = 0

func (q *Queue) Init() {
	go func() {
		for {
			if runningTasks < int(q.MaxConcurrentTasks) {
				task, err := q.TaskRepository.NextQueued()
				if err != nil {
					q.Sev.Logger().Errorf("QUEUE - failed to receive queued task from db: %v", err)
				} else if task == nil {
					q.Sev.Logger().Debug("QUEUE - no queued tasks found")
				} else {
					go q.processTask(task)
				}
			} else {
				q.Sev.Logger().Debugf("QUEUE - maximum concurrent tasks reached (tasks: %d/%d)", runningTasks, q.MaxConcurrentTasks)
			}
			time.Sleep(1 * time.Second) // Delay of 1 second
		}
	}()
}

func (q *Queue) processTask(task *model.Task) {
	runningTasks++
	defer func() { runningTasks-- }()

	q.updateTaskStatus(task, dto.RUNNING)
	q.Sev.Logger().Infof("QUEUE - processing task (uuid: %s)", task.Uuid)

	cmd := task.Command
	err := ffmpeg.Execute(&ffmpeg.ExceutionRequest{Task: task, Command: cmd, InputFile: task.InputFile, OutputFile: task.OutputFile, Logger: q.Sev.Logger()}, func(progress float64) {
		q.TaskRepository.SetTaskProgress(task, progress)
	})
	if err != nil {
		q.updateTaskStatus(task, dto.DONE_ERROR)
		q.Sev.Logger().Warnf("QUEUE - task failed (uuid: %s): %v", task.Uuid, err)
		return
	}

	q.updateTaskStatus(task, dto.DONE_SUCCESSFUL)
	q.Sev.Logger().Infof("QUEUE - task successful (uuid: %s)", task.Uuid)
}

func (q *Queue) updateTaskStatus(task *model.Task, status dto.TaskStatus) {
	q.TaskRepository.SetTaskStatus(task, status)
	q.Sev.Metrics().Gauge("task.status.updated").Inc()
	q.WebhookService.Fire(dto.TASK_STATUS_UPDATED, task.ToDto())
}
