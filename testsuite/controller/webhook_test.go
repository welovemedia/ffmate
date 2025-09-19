package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/service"
	webhookSvc "github.com/welovemedia/ffmate/internal/service/webhook"
	"github.com/welovemedia/ffmate/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/util/testutil"
)

var newWebhook = &dto.Webhook{
	Event: dto.TASK_CREATED,
	Url:   "https://example.com",
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
	webhook, _ := testsuite.ParseJsonBody[dto.Webhook](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/webhooks")
	assert.Equal(t, webhook.Event, dto.TASK_CREATED, "POST /api/v1/webhooks")
	assert.NotEmpty(t, webhook.Uuid, "POST /api/v1/webhooks")
}

func TestWebhookList(t *testing.T) {
	server := testsuite.InitServer(t)

	createWebhook(t, server)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks", nil)
	response := server.TestRequest(request)

	webhooks, _ := testsuite.ParseJsonBody[[]dto.Webhook](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/webhooks")
	assert.Equal(t, 1, len(webhooks), "GET /api/v1/webhooks")
}

func TestWebhookDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	// create webhook
	response := createWebhook(t, server)
	webhook, _ := testsuite.ParseJsonBody[dto.Webhook](response.Body)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/webhooks/"+webhook.Uuid, nil)
	response = server.TestRequest(request)
	assert.Equal(t, 204, response.StatusCode, "DELETE /api/v1/webhooks")

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/webhooks/"+webhook.Uuid, nil)
	response = server.TestRequest(request)
	assert.Equal(t, 400, response.StatusCode, "DELETE /api/v1/webhooks")
}

func TestWebhookGet(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWebhook(t, server)
	webhook, _ := testsuite.ParseJsonBody[dto.Webhook](response.Body)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/"+webhook.Uuid, nil)
	response = server.TestRequest(request)
	webhook, _ = testsuite.ParseJsonBody[dto.Webhook](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/webhooks/{uuid}")
	assert.Equal(t, webhook.Event, dto.TASK_CREATED, "GET /api/v1/webhooks/{uuid}")
}

func TestWebhookUpdate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWebhook(t, server)
	webhook, _ := testsuite.ParseJsonBody[dto.Webhook](response.Body)

	webhook.Event = dto.TASK_UPDATED
	body, _ := json.Marshal(webhook)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response = server.TestRequest(request)
	webhook, _ = testsuite.ParseJsonBody[dto.Webhook](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/webhooks/{uuid}")
	assert.Equal(t, webhook.Event, dto.TASK_UPDATED, "GET /api/v1/webhooks/{uuid}")
}

func TestWebhookDelivery(t *testing.T) {
	server := testsuite.InitServer(t)

	done := make(chan struct{})

	var webhookServer *httptest.Server

	webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := testsuite.ParseJsonBody[struct {
			Event dto.WebhookEvent `json:"event"`
			Data  dto.Webhook      `json:"data"`
		}](r.Body)
		assert.Equal(t, payload.Event, dto.TASK_CREATED, "RECV Webhook")
		assert.Equal(t, payload.Data.Url, webhookServer.URL, "RECV Webhook")
		w.WriteHeader(http.StatusNoContent)

		close(done)
	}))
	defer webhookServer.Close()

	nw := &dto.Webhook{
		Event: dto.TASK_CREATED,
		Url:   webhookServer.URL,
	}
	body, _ := json.Marshal(nw)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	server.TestRequest(request)

	svc := server.Service(service.Webhook).(*webhookSvc.Service)
	svc.Fire(dto.TASK_CREATED, nw)

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
		payload, _ := testsuite.ParseJsonBody[struct {
			Event dto.WebhookEvent `json:"event"`
			Data  dto.Preset       `json:"data"`
		}](r.Body)
		assert.Equal(t, payload.Event, dto.PRESET_CREATED, "RECV Webhook")
		assert.Equal(t, (*payload.Data.Webhooks)[0].Url, webhookServer.URL, "RECV Webhook")
		assert.Equal(t, payload.Data.Name, "Test", "RECV Webhook")
		w.WriteHeader(http.StatusNoContent)

		close(done)
	}))
	defer webhookServer.Close()

	np := &dto.Preset{
		Name:    "Test",
		Command: "-y",
		Webhooks: &dto.DirectWebhooks{
			dto.NewWebhook{
				Event: dto.PRESET_CREATED,
				Url:   webhookServer.URL,
			},
		},
	}
	body, _ := json.Marshal(np)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/presets", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	server.TestRequest(request)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook was never delivered (timeout)")
	}
}
