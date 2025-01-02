package queue

import (
	"time"

	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/pkg/ffmpeg"
	"github.com/welovemedia/ffmate/sev"
)

type Queue struct {
	Sev                *sev.Sev
	TaskRepository     *repository.Task
	MaxConcurrentTasks uint
}

var runningTasks = 0

func (q *Queue) Init() {
	go func() {
		for {
			if runningTasks < int(q.MaxConcurrentTasks) {
				task, err := q.TaskRepository.NextQueued()
				if err != nil {
					q.Sev.Logger().Errorf("failed to receive queued task from db: %v", err)
				} else if task == nil {
					q.Sev.Logger().Debug("no queued tasks found")
				} else {
					go q.processTask(task)
				}
			} else {
				q.Sev.Logger().Debugf("maximum concurrent tasks reached (tasks: %d/%d)", runningTasks, q.MaxConcurrentTasks)
			}
			time.Sleep(1 * time.Second) // Delay of 1 second
		}
	}()
}

func (q *Queue) processTask(task *model.Task) {
	runningTasks++
	q.Sev.Logger().Infof("processing task (uuid: %s)", task.Uuid)
	q.TaskRepository.SetTaskStatus(task, dto.RUNNING)
	cmd := task.Command
	err := ffmpeg.Execute(&ffmpeg.ExceutionRequest{Command: cmd, InputFile: task.InputFile, OutputFile: task.OutputFile})
	if err != nil {
		q.TaskRepository.SetTaskStatus(task, dto.DONE_ERROR)
		q.Sev.Logger().Warnf("task failed (uuid: %s): %v", task.Uuid, err)
		runningTasks--
		return
	}
	q.TaskRepository.SetTaskStatus(task, dto.DONE_SUCCESSFUL)
	q.Sev.Logger().Infof("task successful (uuid: %s)", task.Uuid)
	runningTasks--
}
