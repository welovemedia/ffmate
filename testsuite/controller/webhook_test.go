package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/service"
	webhookSvc "github.com/welovemedia/ffmate/v2/internal/service/webhook"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/util/testutil"
)

var newWebhook = &dto.Webhook{
	Event: dto.TaskCreated,
	URL:   "https://example.com",
}

func createWebhook(t *testing.T, server *testutil.TestServer) *http.Response {
	body, _ := json.Marshal(newWebhook)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/webhooks")
	return response
}

func TestWebhookCreate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWebhook(t, server)
	defer response.Body.Close() // nolint:errcheck
	webhook, _ := testsuite.ParseJSONBody[dto.Webhook](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/webhooks")
	assert.Equal(t, dto.TaskCreated, webhook.Event, "POST /api/v1/webhooks")
	assert.NotEmpty(t, webhook.UUID, "POST /api/v1/webhooks")
}

func TestWebhookList(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWebhook(t, server)
	defer response.Body.Close() // nolint:errcheck

	request := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	webhooks, _ := testsuite.ParseJSONBody[[]dto.Webhook](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/webhooks")
	assert.Len(t, webhooks, 1, "GET /api/v1/webhooks")
}

func TestWebhookDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	// create webhook
	response := createWebhook(t, server)
	defer response.Body.Close() // nolint:errcheck
	webhook, _ := testsuite.ParseJSONBody[dto.Webhook](response.Body)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/webhooks/"+webhook.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 204, response.StatusCode, "DELETE /api/v1/webhooks")

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/webhooks/"+webhook.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 400, response.StatusCode, "DELETE /api/v1/webhooks")
}

func TestWebhookGet(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWebhook(t, server)
	defer response.Body.Close() // nolint:errcheck
	webhook, _ := testsuite.ParseJSONBody[dto.Webhook](response.Body)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/"+webhook.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	webhook, _ = testsuite.ParseJSONBody[dto.Webhook](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/webhooks/{uuid}")
	assert.Equal(t, dto.TaskCreated, webhook.Event, "GET /api/v1/webhooks/{uuid}")
}

func TestWebhookUpdate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWebhook(t, server)
	defer response.Body.Close() // nolint:errcheck
	webhook, _ := testsuite.ParseJSONBody[dto.Webhook](response.Body)

	webhook.Event = dto.TaskUpdated
	body, _ := json.Marshal(webhook)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	webhook, _ = testsuite.ParseJSONBody[dto.Webhook](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/webhooks/{uuid}")
	assert.Equal(t, dto.TaskUpdated, webhook.Event, "GET /api/v1/webhooks/{uuid}")
}

func TestWebhookDelivery(t *testing.T) {
	server := testsuite.InitServer(t)

	done := make(chan struct{})

	var webhookServer *httptest.Server

	webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := testsuite.ParseJSONBody[struct {
			Data  dto.Webhook      `json:"data"`
			Event dto.WebhookEvent `json:"event"`
		}](r.Body)
		assert.Equal(t, dto.TaskCreated, payload.Event, "RECV Webhook")
		assert.Equal(t, payload.Data.URL, webhookServer.URL, "RECV Webhook")
		w.WriteHeader(http.StatusNoContent)

		close(done)
	}))
	defer webhookServer.Close() // nolint:errcheck

	nw := &dto.Webhook{
		Event: dto.TaskCreated,
		URL:   webhookServer.URL,
	}
	body, _ := json.Marshal(nw)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	resp := server.TestRequest(request)
	defer resp.Body.Close() // nolint:errcheck

	svc := server.Service(service.Webhook).(*webhookSvc.Service)
	svc.Fire(dto.TaskCreated, nw)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook was never delivered (timeout)")
	}
}

func TestWebhookDirectDelivery(t *testing.T) {
	server := testsuite.InitServer(t)

	done := make(chan struct{})

	var webhookServer *httptest.Server
	webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := testsuite.ParseJSONBody[struct {
			Event dto.WebhookEvent `json:"event"`
			Data  dto.Preset       `json:"data"`
		}](r.Body)
		assert.Equal(t, dto.PresetCreated, payload.Event, "RECV Webhook")
		assert.Equal(t, (*payload.Data.Webhooks)[0].URL, webhookServer.URL, "RECV Webhook")
		assert.Equal(t, "Test", payload.Data.Name, "RECV Webhook")
		w.WriteHeader(http.StatusNoContent)

		close(done)
	}))
	defer webhookServer.Close() // nolint:errcheck

	np := &dto.Preset{
		Name:    "Test",
		Command: "-y",
		Webhooks: &dto.DirectWebhooks{
			dto.NewWebhook{
				Event: dto.PresetCreated,
				URL:   webhookServer.URL,
			},
		},
	}
	body, _ := json.Marshal(np)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/presets", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	resp := server.TestRequest(request)
	defer resp.Body.Close() // nolint:errcheck

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook was never delivered (timeout)")
	}
}
