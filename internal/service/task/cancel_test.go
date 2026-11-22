package task

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

func TestCancelRunningTaskNotifiesOnce(t *testing.T) {
	svc, taskRepo, _, rec := newServiceWithWebhooks(t, dto.TaskUpdated, dto.BatchFinished)

	task, err := taskRepo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.Running, Batch: uuid.NewString()})
	assert.NoError(t, err)

	// API cancel
	_, err = svc.Cancel(task.UUID)
	assert.NoError(t, err)

	// the processing goroutine notices the cancelation (ffmpeg stopped) and finishes the task
	running := *task
	svc.cancelTask(&running)
	// a pre/post-processing script failing after the cancelation must not overwrite it either
	svc.failTask(&running, errors.New("script failed"))
	// canceling again is a no-op
	_, err = svc.Cancel(task.UUID)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		return rec.count(dto.TaskUpdated, task.UUID) >= 1 && rec.count(dto.BatchFinished, task.UUID) >= 1
	}, 5*time.Second, 50*time.Millisecond)
	time.Sleep(500 * time.Millisecond) // give duplicate deliveries time to arrive

	assert.Equal(t, 1, rec.count(dto.TaskUpdated, task.UUID), "one task.updated per cancel")
	assert.Equal(t, 1, rec.count(dto.BatchFinished, task.UUID), "one batch.finished per batch")
	for _, e := range rec.all() {
		assert.Equal(t, dto.DoneCanceled, e.Data.Status)
	}

	stored, _ := taskRepo.First(task.UUID)
	assert.Equal(t, dto.DoneCanceled, stored.Status, "the canceled state must not be overwritten")
	assert.Empty(t, stored.Error)
}

func TestFailTaskStillNotifiesWhenNotCanceled(t *testing.T) {
	svc, taskRepo, _, rec := newServiceWithWebhooks(t, dto.TaskUpdated)

	task, err := taskRepo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.Running})
	assert.NoError(t, err)
	svc.failTask(task, errors.New("boom"))

	assert.Eventually(t, func() bool { return rec.count(dto.TaskUpdated, task.UUID) == 1 }, 5*time.Second, 50*time.Millisecond)
	stored, _ := taskRepo.First(task.UUID)
	assert.Equal(t, dto.DoneError, stored.Status)
	assert.Equal(t, "boom", stored.Error)
}
