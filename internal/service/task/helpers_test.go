package task

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

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
	"gorm.io/gorm"
)

// webhookEvent is a webhook request received by the test receiver
type webhookEvent struct {
	Event string   `json:"event"`
	Data  dto.Task `json:"data"`
}

type webhookReceiver struct {
	mu     sync.Mutex
	events []webhookEvent
}

func (r *webhookReceiver) all() []webhookEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]webhookEvent(nil), r.events...)
}

// count returns the number of received events of the given type for the given task
func (r *webhookReceiver) count(event dto.WebhookEvent, taskUUID string) int {
	n := 0
	for _, e := range r.all() {
		if e.Event == string(event) && e.Data.UUID == taskUUID {
			n++
		}
	}
	return n
}

// newServiceWithWebhooks creates a task service on an in-memory database with global webhooks
// for the given events pointing to a local receiver
func newServiceWithWebhooks(t *testing.T, events ...dto.WebhookEvent) (*Service, *repository.Task, *gorm.DB, *webhookReceiver) {
	t.Helper()
	server := testserver.New(t)
	db := server.DB()
	(&repository.Client{DB: db}).Setup()
	taskRepo := (&repository.Task{DB: db}).Setup()
	webhookRepo := (&repository.Webhook{DB: db}).Setup()
	websocketSvc := websocketService.NewService(db)
	webhookSvc := webhookService.NewService(webhookRepo, (&repository.WebhookExecution{DB: db}).Setup(), server.Config(), websocketSvc)
	presetSvc := presetService.NewService((&repository.Preset{DB: db}).Setup(), webhookSvc, websocketSvc)
	svc := &Service{repository: taskRepo, presetService: presetSvc, webhookService: webhookSvc, websocketService: websocketSvc, ffmpegService: ffmpeg.NewService()}

	rec := &webhookReceiver{}
	receiver := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var e webhookEvent
		_ = json.Unmarshal(b, &e)
		rec.mu.Lock()
		rec.events = append(rec.events, e)
		rec.mu.Unlock()
	}))
	t.Cleanup(receiver.Close)
	for _, event := range events {
		_, err := webhookRepo.Add(&model.Webhook{UUID: uuid.NewString(), Event: event, URL: receiver.URL})
		assert.NoError(t, err)
	}
	return svc, taskRepo, db, rec
}
