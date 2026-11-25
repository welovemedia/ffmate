package task

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
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

func TestCancelStopsParallelRunningTasks(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	cfg.Set("ffmate.ffmpeg", ffmpegPath)
	svc, taskRepo, _, _ := newServiceWithWebhooks(t)
	dir := t.TempDir()

	const n = 30
	tasks := make([]*model.Task, n)
	var wg sync.WaitGroup
	// never leave ffmpeg processes behind, even if the test fails
	t.Cleanup(func() {
		for _, task := range tasks {
			if task == nil {
				continue
			}
			if ctxVal, ok := taskQueue.Load(task.UUID); ok {
				ctxVal.(taskContext).cancel(errors.New("test cleanup"))
			}
		}
		wg.Wait()
	})
	for i := range tasks {
		task, err := taskRepo.Add(&model.Task{
			UUID:   uuid.NewString(),
			Status: dto.Running,
			// -stats_period 30: without progress updates, only a direct context cancel stops ffmpeg in time
			Command:    &dto.RawResolved{Raw: "-re -f lavfi -i testsrc=d=60:s=64x64 -stats_period 30 -y ${OUTPUT_FILE}"},
			InputFile:  &dto.RawResolved{},
			OutputFile: &dto.RawResolved{Raw: filepath.Join(dir, fmt.Sprintf("out-%d.mp4", i))},
		})
		assert.NoError(t, err)
		tasks[i] = task
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.processNewTask(task)
		}()
	}

	// wait until all tasks are running and ffmpeg has started
	assert.Eventually(t, func() bool {
		for _, task := range tasks {
			stored, _ := taskRepo.First(task.UUID)
			if stored == nil || stored.Status != dto.Running || stored.Command.Resolved == "" {
				return false
			}
		}
		return true
	}, 20*time.Second, 200*time.Millisecond, "all tasks running")
	time.Sleep(time.Second)

	// cancel all tasks at the same time
	var cancels sync.WaitGroup
	for _, task := range tasks {
		cancels.Add(1)
		go func() {
			defer cancels.Done()
			_, err := svc.Cancel(task.UUID)
			assert.NoError(t, err)
		}()
	}
	cancels.Wait()

	// processing must stop right away (not at the end of the 60s source)
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("canceled tasks are still being processed")
	}

	for _, task := range tasks {
		stored, _ := taskRepo.First(task.UUID)
		assert.Equal(t, dto.DoneCanceled, stored.Status, "task %s must stay canceled", task.UUID)
	}
}
