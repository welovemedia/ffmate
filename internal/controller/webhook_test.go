package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/metrics"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/sev"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type WebhookCall struct {
	event dto.WebhookEvent
	data  interface{}
}

func setupWebhookServer(t *testing.T, expectedCalls int) (string, chan WebhookCall) {
	calls := make(chan WebhookCall, expectedCalls)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		calls <- WebhookCall{
			event: dto.WebhookEvent(payload["event"].(string)),
			data:  payload["data"],
		}
		w.WriteHeader(http.StatusOK)
	}))

	return server.URL, calls
}

func waitForWebhook(t *testing.T, calls chan WebhookCall, expectedEvent dto.WebhookEvent) {
	select {
	case call := <-calls:
		if call.event != expectedEvent {
			t.Errorf("Expected webhook event %s, got %s", expectedEvent, call.event)
		}
	case <-time.After(time.Second):
		t.Errorf("Timeout waiting for webhook event %s", expectedEvent)
	}
}

func setupWebhookTestDB(t *testing.T) (*gorm.DB, *sev.Sev) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&model.Webhook{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	s := sev.New("test", "", "", 3000)
	s.SetDB(db)
	service.Init(s)

	metrics := &metrics.Metrics{}
	for name, gauge := range metrics.Gauges() {
		s.Metrics().RegisterGauge(name, gauge)
	}
	for name, gauge := range metrics.GaugesVec() {
		s.Metrics().RegisterGaugeVec(name, gauge)
	}

	return db, s
}

func TestWebhookController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, s := setupWebhookTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying database: %v", err)
	}
	defer sqlDB.Close()

	// Setup webhook server and expected calls
	webhookURL, webhookCalls := setupWebhookServer(t, 3) // Expect: created, updated, deleted

	// Setup webhooks for webhook events
	service.WebhookService().NewWebhook(&dto.NewWebhook{
		Event: dto.WEBHOOK_CREATED,
		Url:   webhookURL,
	})
	service.WebhookService().NewWebhook(&dto.NewWebhook{
		Event: dto.WEBHOOK_UPDATED,
		Url:   webhookURL,
	})
	service.WebhookService().NewWebhook(&dto.NewWebhook{
		Event: dto.WEBHOOK_DELETED,
		Url:   webhookURL,
	})

	// Wait for the webhook listener creation events (as we create two webhook listeners)
	waitForWebhook(t, webhookCalls, dto.WEBHOOK_CREATED)
	waitForWebhook(t, webhookCalls, dto.WEBHOOK_CREATED)
	waitForWebhook(t, webhookCalls, dto.WEBHOOK_CREATED)

	controller := &WebhookController{
		Prefix: "",
	}
	controller.Setup(s)

	t.Run("Add webhook", func(t *testing.T) {
		newWebhook := dto.NewWebhook{
			Event: dto.TASK_CREATED,
			Url:   "http://localhost:8080/test",
		}

		body, _ := json.Marshal(newWebhook)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/webhooks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		s.Gin().ServeHTTP(w, req)
		waitForWebhook(t, webhookCalls, dto.WEBHOOK_CREATED)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response dto.Webhook
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Event != newWebhook.Event {
			t.Errorf("Expected event %s, got %s", newWebhook.Event, response.Event)
		}
	})

	t.Run("List webhooks", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/v1/webhooks?page=0&perPage=10", nil)
		s.Gin().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response []dto.Webhook
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response) < 1 {
			t.Error("Expected at least one webhook")
		}
	})

	t.Run("Get webhook", func(t *testing.T) {
		webhooks, _, _ := service.WebhookService().ListWebhooks(0, 1)
		if len(*webhooks) > 0 {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/v1/webhooks/"+(*webhooks)[0].Uuid, nil)
			s.Gin().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response dto.Webhook
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}
	})

	t.Run("Update webhook", func(t *testing.T) {
		webhooks, _, _ := service.WebhookService().ListWebhooks(0, 1)
		if len(*webhooks) > 0 {
			newWebhook := dto.NewWebhook{
				Event: dto.TASK_UPDATED,
				Url:   "http://localhost:8080/test",
			}

			body, _ := json.Marshal(newWebhook)
			w := httptest.NewRecorder()
			uuid := (*webhooks)[0].Uuid
			req := httptest.NewRequest("PUT", "/v1/webhooks/"+uuid, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			s.Gin().ServeHTTP(w, req)
			waitForWebhook(t, webhookCalls, dto.WEBHOOK_UPDATED)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response dto.Webhook
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Event != newWebhook.Event {
				t.Errorf("Expected event %s, got %s", newWebhook.Event, response.Event)
			}
		}
	})

	t.Run("Delete webhook", func(t *testing.T) {
		webhooks, _, _ := service.WebhookService().ListWebhooks(0, 1)
		if len(*webhooks) > 0 {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", "/v1/webhooks/"+(*webhooks)[0].Uuid, nil)
			s.Gin().ServeHTTP(w, req)
			waitForWebhook(t, webhookCalls, dto.WEBHOOK_DELETED)

			if w.Code != http.StatusNoContent {
				t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
			}
		}
	})
}
