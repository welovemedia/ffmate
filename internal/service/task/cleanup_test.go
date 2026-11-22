package task

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/service/ffmpeg"
	presetService "github.com/welovemedia/ffmate/v2/internal/service/preset"
	webhookService "github.com/welovemedia/ffmate/v2/internal/service/webhook"
	websocketService "github.com/welovemedia/ffmate/v2/internal/service/websocket"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
)

func TestCleanupNotifiesAndRetries(t *testing.T) {
	server := testserver.New(t)
	db := server.DB()
	(&repository.Client{DB: db}).Setup()
	taskRepo := (&repository.Task{DB: db}).Setup()
	webhookRepo := (&repository.Webhook{DB: db}).Setup()
	websocketSvc := websocketService.NewService(db)
	webhookSvc := webhookService.NewService(webhookRepo, (&repository.WebhookExecution{DB: db}).Setup(), server.Config(), websocketSvc)
	presetSvc := presetService.NewService((&repository.Preset{DB: db}).Setup(), webhookSvc, websocketSvc)
	svc := &Service{repository: taskRepo, presetService: presetSvc, webhookService: webhookSvc, websocketService: websocketSvc, ffmpegService: ffmpeg.NewService()}

	// receiver for the global task.updated webhook
	var mu sync.Mutex
	var bodies []string
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		mu.Unlock()
	}))
	defer receiver.Close()
	_, err := webhookRepo.Add(&model.Webhook{UUID: uuid.NewString(), Event: dto.TaskUpdated, URL: receiver.URL})
	assert.NoError(t, err)

	assert.NoError(t, db.Create(&model.Client{Identifier: "crashed"}).Error)
	withRetries, err := taskRepo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.PreProcessing, ClientIdentifier: "crashed", Retries: 1})
	assert.NoError(t, err)
	noRetries, err := taskRepo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.PostProcessing, ClientIdentifier: "crashed"})
	assert.NoError(t, err)

	tasks, err := taskRepo.FailRunningTasksForStartingClient("crashed")
	assert.NoError(t, err)
	svc.notifyCleanedUp(tasks)

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		n := 0
		for _, b := range bodies {
			if strings.Contains(b, `"status":"DONE_ERROR"`) && (strings.Contains(b, withRetries.UUID) || strings.Contains(b, noRetries.UUID)) {
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
