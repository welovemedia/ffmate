package task

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

func TestCleanupNotifiesAndRetries(t *testing.T) {
	svc, taskRepo, db, rec := newServiceWithWebhooks(t, dto.TaskUpdated)

	assert.NoError(t, db.Create(&model.Client{Identifier: "crashed"}).Error)
	withRetries, err := taskRepo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.PreProcessing, ClientIdentifier: "crashed", Retries: 1})
	assert.NoError(t, err)
	noRetries, err := taskRepo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.PostProcessing, ClientIdentifier: "crashed"})
	assert.NoError(t, err)

	tasks, err := taskRepo.FailRunningTasksForStartingClient("crashed")
	assert.NoError(t, err)
	svc.notifyCleanedUp(tasks)

	assert.Eventually(t, func() bool {
		n := 0
		for _, e := range rec.all() {
			if e.Event == string(dto.TaskUpdated) && e.Data.Status == dto.DoneError && (e.Data.UUID == withRetries.UUID || e.Data.UUID == noRetries.UUID) {
				n++
			}
		}
		return n == 2
	}, 5*time.Second, 50*time.Millisecond, "task.updated webhook must be sent for each cleaned up task")

	// a task with retries left is re-queued (after the regular 3s retry delay)
	assert.Eventually(t, func() bool {
		task, _ := taskRepo.First(withRetries.UUID)
		return task != nil && task.Status == dto.Queued && task.Retried == 1
	}, 6*time.Second, 100*time.Millisecond, "task with retries must be re-queued")

	task, _ := taskRepo.First(noRetries.UUID)
	assert.Equal(t, dto.DoneError, task.Status, "task without retries stays failed")
}
