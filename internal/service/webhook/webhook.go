package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
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
	Add(*model.Webhook) (*model.Webhook, error)
	Update(*model.Webhook) (*model.Webhook, error)
	First(string) (*model.Webhook, error)
	Delete(*model.Webhook) error
	Count() (int64, error)
}

type ExecutionRepository interface {
	List(page int, perPage int) (*[]model.WebhookExecution, int64, error)
	Add(*model.WebhookExecution) (*model.WebhookExecution, error)
	Count() (int64, error)
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
	w, err := s.repository.Add(&model.Webhook{Uuid: uuid.NewString(), Event: newWebhook.Event, Url: newWebhook.Url})
	debug.Log.Info("created webhook (uuid: %s)", w.Uuid)

	metrics.Gauge("webhook.created").Inc()
	s.Fire(dto.WEBHOOK_CREATED, w.ToDto())
	s.websocketService.Broadcast(websocket.WEBHOOK_CREATED, w.ToDto())

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
	w.Url = newWebhook.Url

	w, err = s.repository.Update(w)
	if err != nil {
		debug.Log.Error("failed to update webhook (uuid: %s): %v", w.Uuid, err)
		return nil, err
	}

	debug.Log.Info("updated webhook (uuid: %s)", w.Uuid)

	metrics.Gauge("webhook.updated").Inc()
	s.Fire(dto.WEBHOOK_UPDATED, w.ToDto())
	s.websocketService.Broadcast(websocket.WEBHOOK_UPDATED, w.ToDto())

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
	s.Fire(dto.WEBHOOK_DELETED, w.ToDto())
	s.websocketService.Broadcast(websocket.WEBHOOK_DELETED, w.ToDto())

	return nil
}

func (s *Service) Fire(event dto.WebhookEvent, data any) error {
	webhooks, err := s.repository.ListAllByEvent(event)
	for _, webhook := range *webhooks {
		go s.fireWebhook(&webhook, data, s.handleWebhookExecution)
		metrics.Gauge("webhook.executed").Inc()
	}
	return err
}

func (s *Service) FireDirect(webhooks *dto.DirectWebhooks, event dto.WebhookEvent, data any) {
	if webhooks == nil {
		return
	}
	for _, webhook := range *webhooks {
		if webhook.Event == event {
			go s.fireWebhook(&model.Webhook{Uuid: uuid.NewString(), Event: webhook.Event, Url: webhook.Url}, data, s.handleWebhookExecution)
			metrics.Gauge("webhook.executed.direct").Inc()
		}
	}
}

func (s *Service) handleWebhookExecution(event dto.WebhookEvent, url string, req *http.Request, resp *http.Response) {
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
		Uuid:     uuid.NewString(),
		Event:    event,
		Url:      url,
		Request:  request,
		Response: response,
	})
	if err != nil {
		debug.Log.Error("failed to create webhook execution for event '%s': %v", event, err)
	} else {
		debug.Webhook.Debug("created new webhook execution for event '%s'", event)
		s.websocketService.Broadcast(websocket.WEBHOOK_EXECUTION_CREATED, w.ToDto())
	}
}

func (s *Service) fireWebhook(webhook *model.Webhook, data any, callback func(event dto.WebhookEvent, url string, req *http.Request, resp *http.Response)) {
	msg := map[string]interface{}{
		"event": webhook.Event,
		"data":  data,
	}
	b, err := json.Marshal(&msg)
	if err != nil {
		debug.Log.Error("failed to fire webhook due to marshalling for event '%s' (uuid: %s): %v", webhook.Event, webhook.Uuid, err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", webhook.Url, bytes.NewBuffer(b))
	if err != nil {
		debug.Log.Error("failed to create http request", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", s.config.GetString("app.name")+"/"+s.config.GetString("app.version"))

	retryDelays := []time.Duration{
		3 * time.Second,
		5 * time.Second,
		10 * time.Second,
	}

	var resp *http.Response
	for try := 0; try <= len(retryDelays); try++ {
		resp, err = client.Do(req)
		if err == nil {
			debug.Webhook.Debug("fired webhook for event '%s' (uuid: %s)", webhook.Event, webhook.Uuid)
			break
		}

		if try < len(retryDelays) {
			time.Sleep(retryDelays[try])
			continue
		}

		debug.Log.Error("failed to fire webhook for event '%s' (uuid: %s) after %d tries", webhook.Event, webhook.Uuid, try+1)
	}

	// Save a copy of body for the callback
	req.Body = io.NopCloser(bytes.NewBuffer(b))
	callback(webhook.Event, webhook.Url, req, resp)
}

func (s *Service) Name() string {
	return service.Webhook
}
