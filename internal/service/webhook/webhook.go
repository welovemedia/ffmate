package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/websocket"
	"goyave.dev/goyave/v5/config"
)

type Repository interface {
	List(page int, perPage int) (*[]model.Webhook, int64, error)
	ListAllByEvent(event dto.WebhookEvent) (*[]model.Webhook, error)
	Add(webhook *model.Webhook) (*model.Webhook, error)
	Update(webhook *model.Webhook) (*model.Webhook, error)
	First(uuid string) (*model.Webhook, error)
	Delete(webhook *model.Webhook) error
	Count() (int64, error)
}

type ExecutionRepository interface {
	List(page int, perPage int) (*[]model.WebhookExecution, int64, error)
	Add(webhookExecution *model.WebhookExecution) (*model.WebhookExecution, error)
}

type Service struct {
	repository          Repository
	executionRepository ExecutionRepository
	config              *config.Config
	websocketService    *websocket.Service
}

func NewService(repository Repository, executionRepository ExecutionRepository, config *config.Config, websocketService *websocket.Service) *Service {
	return &Service{
		repository:          repository,
		executionRepository: executionRepository,
		config:              config,
		websocketService:    websocketService,
	}
}

func (s *Service) Get(uuid string) (*model.Webhook, error) {
	p, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, errors.New("webhook for given uuid not found")
	}

	return p, nil
}

func (s *Service) List(page int, perPage int) (*[]model.Webhook, int64, error) {
	return s.repository.List(page, perPage)
}

func (s *Service) ListExecutions(page int, perPage int) (*[]model.WebhookExecution, int64, error) {
	return s.executionRepository.List(page, perPage)
}

func (s *Service) Add(newWebhook *dto.NewWebhook) (*model.Webhook, error) {
	w, err := s.repository.Add(&model.Webhook{UUID: uuid.NewString(), Event: newWebhook.Event, URL: newWebhook.URL})
	debug.Log.Info("created webhook (uuid: %s)", w.UUID)

	metrics.Gauge("webhook.created").Inc()
	s.Fire(dto.WebhookCreated, w.ToDTO())
	s.websocketService.Broadcast(websocket.WebhookCreated, w.ToDTO())

	return w, err
}

func (s *Service) Update(uuid string, newWebhook *dto.NewWebhook) (*model.Webhook, error) {
	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("webhook for given uuid not found")
	}

	w.Event = newWebhook.Event
	w.URL = newWebhook.URL

	w, err = s.repository.Update(w)
	if err != nil {
		debug.Log.Error("failed to update webhook (uuid: %s): %v", w.UUID, err)
		return nil, err
	}

	debug.Log.Info("updated webhook (uuid: %s)", w.UUID)

	metrics.Gauge("webhook.updated").Inc()
	s.Fire(dto.WebhookUpdated, w.ToDTO())
	s.websocketService.Broadcast(websocket.WebhookUpdated, w.ToDTO())

	return w, err
}

func (s *Service) Delete(uuid string) error {
	w, err := s.repository.First(uuid)
	if err != nil {
		return err
	}

	if w == nil {
		return errors.New("webhook for given uuid not found")
	}

	err = s.repository.Delete(w)
	if err != nil {
		debug.Log.Error("failed to delete webhook (uuid: %s)", uuid)
		return err
	}

	debug.Log.Info("deleted webhook (uuid: %s)", uuid)

	metrics.Gauge("webhook.deleted").Inc()
	s.Fire(dto.WebhookDeleted, w.ToDTO())
	s.websocketService.Broadcast(websocket.WebhookDeleted, w.ToDTO())

	return nil
}

func (s *Service) Fire(event dto.WebhookEvent, data any) {
	webhooks, _ := s.repository.ListAllByEvent(event)
	for _, webhook := range *webhooks {
		go s.fireWebhook(&webhook, data, s.handleWebhookExecution)
		metrics.Gauge("webhook.executed").Inc()
	}
}

// FireInRoutine will execute in the same routine as the caller (mainly used for testing)
func (s *Service) FireInRoutine(event dto.WebhookEvent, data any) {
	webhooks, _ := s.repository.ListAllByEvent(event)
	for _, webhook := range *webhooks {
		s.fireWebhook(&webhook, data, s.handleWebhookExecution)
		metrics.Gauge("webhook.executed").Inc()
	}
}

func (s *Service) FireDirect(webhooks *dto.DirectWebhooks, event dto.WebhookEvent, data any) {
	if webhooks == nil {
		return
	}
	for _, webhook := range *webhooks {
		if webhook.Event == event {
			go s.fireWebhook(&model.Webhook{UUID: uuid.NewString(), Event: webhook.Event, URL: webhook.URL}, data, s.handleWebhookExecution)
			metrics.Gauge("webhook.executed.direct").Inc()
		}
	}
}

func (s *Service) handleWebhookExecution(event dto.WebhookEvent, url string, req *http.Request, resp *http.Response) {
	if req == nil || resp == nil {
		return
	}

	// Read request body
	var reqBodyStr string
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		reqBodyStr = string(bodyBytes)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Read response body
	var respBodyStr string
	if resp.Body != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		respBodyStr = string(bodyBytes)
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	request := &dto.WebhookRequest{
		Headers: req.Header,
		Body:    reqBodyStr,
	}

	response := &dto.WebhookResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Body:    respBodyStr,
	}

	w, err := s.executionRepository.Add(&model.WebhookExecution{
		UUID:     uuid.NewString(),
		Event:    event,
		URL:      url,
		Request:  request,
		Response: response,
	})
	if err != nil {
		debug.Log.Error("failed to create webhook execution for event '%s': %v", event, err)
	} else {
		debug.Webhook.Debug("created new webhook execution for event '%s'", event)
		s.websocketService.Broadcast(websocket.WebhookExecutionCreated, w.ToDTO())
	}
}

// webhookPayload is the body sent to webhook receivers (field order as documented)
type webhookPayload struct {
	Event     dto.WebhookEvent `json:"event"`
	Timestamp string           `json:"timestamp"` // RFC 3339, UTC
	Data      any              `json:"data"`
}

// webhookClient is shared by all webhook deliveries; the timeout limits each single try
var webhookClient = &http.Client{Timeout: 10 * time.Second}

// webhookRetryDelays defines the delays between delivery tries (after a request error, a timeout or a non-2xx response)
var webhookRetryDelays = []time.Duration{
	3 * time.Second,
	5 * time.Second,
	10 * time.Second,
}

func (s *Service) fireWebhook(webhook *model.Webhook, data any, callback func(event dto.WebhookEvent, url string, req *http.Request, resp *http.Response)) {
	// fireWebhook runs in its own goroutine, a panic must not take down the server
	defer func() {
		if r := recover(); r != nil {
			debug.Log.Error("recovered from panic while firing webhook for event '%s' (uuid: %s): %v", webhook.Event, webhook.UUID, r)
		}
	}()

	// the timestamp is set once per event, so all retries carry the same value
	b, err := json.Marshal(&webhookPayload{
		Event:     webhook.Event,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	})
	if err != nil {
		debug.Log.Error("failed to fire webhook due to marshaling for event '%s' (uuid: %s): %v", webhook.Event, webhook.UUID, err)
		return
	}

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(b))
	if err != nil {
		debug.Log.Error("failed to create http request for webhook (uuid: %s): %v", webhook.UUID, err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", s.config.GetString("app.name")+"/"+s.config.GetString("app.version"))

	tries := len(webhookRetryDelays) + 1
	for try := 0; try < tries; try++ {
		if try > 0 {
			time.Sleep(webhookRetryDelays[try-1])
		}

		// reset the body as it has been consumed by a previous try
		req.Body = io.NopCloser(bytes.NewBuffer(b))
		resp, err := webhookClient.Do(req)
		if err == nil {
			// record every received response (also non-2xx) as webhook execution
			req.Body = io.NopCloser(bytes.NewBuffer(b))
			callback(webhook.Event, webhook.URL, req, resp)
			_ = resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				debug.Webhook.Debug("fired webhook for event '%s' (uuid: %s)", webhook.Event, webhook.UUID)
				return
			}
			err = fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}

		debug.Webhook.Warn("webhook delivery for event '%s' failed (uuid: %s, try %d/%d): %v", webhook.Event, webhook.UUID, try+1, tries, err)
	}

	debug.Log.Error("failed to fire webhook for event '%s' (uuid: %s) after %d tries", webhook.Event, webhook.UUID, tries)
}

func (s *Service) Name() string {
	return service.Webhook
}
