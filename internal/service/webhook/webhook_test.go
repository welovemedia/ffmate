package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"goyave.dev/goyave/v5/config"
)

func newTestService(t *testing.T) *Service {
	conf, err := config.LoadJSON(`{"app":{"name":"ffmate","version":"test"}}`)
	assert.NoError(t, err)
	return &Service{config: conf}
}

func fastRetries(t *testing.T) {
	orig := webhookRetryDelays
	webhookRetryDelays = []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond}
	t.Cleanup(func() { webhookRetryDelays = orig })
}

type recorder struct {
	mu       sync.Mutex
	statuses []int
}

func (r *recorder) callback(_ dto.WebhookEvent, _ string, _ *http.Request, resp *http.Response) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statuses = append(r.statuses, resp.StatusCode)
}

func TestFireWebhookRetriesOnNon2xx(t *testing.T) {
	fastRetries(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rec := &recorder{}
	newTestService(t).fireWebhook(&model.Webhook{UUID: "u", Event: dto.TaskCreated, URL: srv.URL}, nil, rec.callback)

	assert.Equal(t, int32(3), calls.Load(), "must retry until a 2xx response")
	assert.Equal(t, []int{500, 500, 200}, rec.statuses, "every response is recorded")
}

func TestFireWebhookStopsAfterMaxTries(t *testing.T) {
	fastRetries(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	rec := &recorder{}
	newTestService(t).fireWebhook(&model.Webhook{UUID: "u", Event: dto.TaskCreated, URL: srv.URL}, nil, rec.callback)

	assert.Equal(t, int32(len(webhookRetryDelays)+1), calls.Load())
}

func TestFireWebhookTimeout(t *testing.T) {
	fastRetries(t)
	origClient := webhookClient
	webhookClient = &http.Client{Timeout: 50 * time.Millisecond}
	t.Cleanup(func() { webhookClient = origClient })

	var calls atomic.Int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		select { // never answer within the client timeout
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	rec := &recorder{}
	done := make(chan struct{})
	go func() {
		newTestService(t).fireWebhook(&model.Webhook{UUID: "u", Event: dto.TaskCreated, URL: srv.URL}, nil, rec.callback)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("fireWebhook did not give up on a hanging receiver")
	}
	assert.Equal(t, int32(len(webhookRetryDelays)+1), calls.Load(), "a timeout must be retried")
	assert.Empty(t, rec.statuses, "no response, no execution")
}

func TestHandleWebhookExecutionNilResponse(t *testing.T) {
	assert.NotPanics(t, func() {
		(&Service{}).handleWebhookExecution(dto.TaskCreated, "http://x", nil, nil)
	})
}

func TestFireWebhookPayloadHasTimestamp(t *testing.T) {
	fastRetries(t)
	var mu sync.Mutex
	var bodies [][]byte
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, b)
		mu.Unlock()
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError) // force one retry
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	before := time.Now().UTC().Truncate(time.Second)
	newTestService(t).fireWebhook(&model.Webhook{UUID: "u", Event: dto.TaskCreated, URL: srv.URL}, map[string]any{"uuid": "x"}, (&recorder{}).callback)
	after := time.Now().UTC()

	assert.Len(t, bodies, 2)
	// documented order: event, timestamp, data
	assert.Regexp(t, `^\{"event":"task\.created","timestamp":"[^"]+","data":\{"uuid":"x"\}\}$`, string(bodies[0]))

	var payload struct {
		Event     string         `json:"event"`
		Timestamp string         `json:"timestamp"`
		Data      map[string]any `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(bodies[0], &payload))
	ts, err := time.Parse(time.RFC3339, payload.Timestamp)
	assert.NoError(t, err, "timestamp must be RFC 3339")
	assert.True(t, strings.HasSuffix(payload.Timestamp, "Z"), "timestamp must be UTC")
	assert.False(t, ts.Before(before) || ts.After(after), "timestamp must be the event time")

	assert.Equal(t, string(bodies[0]), string(bodies[1]), "retries must send the identical payload (same timestamp)")
}
