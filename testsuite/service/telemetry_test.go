package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/telemetry"
	"github.com/welovemedia/ffmate/v2/testsuite"
)

func TestTelemetry(t *testing.T) {
	server := testsuite.InitServer(t)
	done := make(chan struct{})

	var webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := testsuite.ParseJSONBody[map[string]any](r.Body)
		assert.InDelta(t, float64(1000), payload["runtimeDuration"], 0, "RECV Telemetry")
		assert.Equal(t, "test-1.0.0", payload["appVersion"], "RECV Telemetry")
		assert.True(t, payload["isShuttingDown"].(bool), "RECV Telemetry")
		assert.True(t, payload["isStartUp"].(bool), "RECV Telemetry")
		w.WriteHeader(http.StatusNoContent)

		close(done)
	}))
	defer webhookServer.Close()

	cfg.Set("ffmate.telemetry.send", true)
	cfg.Set("ffmate.telemetry.url", webhookServer.URL)
	svc := server.Service(service.Telemetry).(*telemetry.Service)
	svc.SendTelemetry(time.Now().Add(-1000*time.Millisecond), true, true)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("telemetry was never delivered (timeout)")
	}
}
