package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/testsuite"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/telemetry"
)

func TestTelemetry(t *testing.T) {
	server := testsuite.InitServer(t)
	done := make(chan struct{})

	var webhookServer *httptest.Server

	webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := testsuite.ParseJsonBody[map[string]any](r.Body)
		assert.Equal(t, payload["runtimeDuration"], float64(1000), "RECV Telemetry")
		assert.Equal(t, payload["appVersion"], "test-1.0.0", "RECV Telemetry")
		assert.Equal(t, payload["isShuttingDown"], true, "RECV Telemetry")
		assert.Equal(t, payload["isStartUp"], true, "RECV Telemetry")
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
