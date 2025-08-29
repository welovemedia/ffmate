package service

import (
	"errors"

	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/sev"
)

type webhookSvc struct {
	service
	sev               *sev.Sev
	webhookRepository *repository.Webhook
}

func (s *webhookSvc) ListWebhooks(page int, perPage int) (*[]model.Webhook, int64, error) {
	return s.webhookRepository.List(page, perPage)
}

func (s *webhookSvc) DeleteWebhook(uuid string) error {
	w, err := s.webhookRepository.First(uuid)
	if err != nil {
		return err
	}

	if w.Uuid == "" {
		return errors.New("webhook for given uuid not found")
	}

	err = s.webhookRepository.Delete(w)
	if err != nil {
		s.sev.Logger().Warnf("failed to delete webhook for event %s (uuid: %s): %+v", w.Event, w.Uuid, err)
		return err
	}

	s.sev.Logger().Infof("deleted webhook for event %s (uuid: %s)", w.Event, w.Uuid)

	s.sev.Metrics().Gauge("webhook.deleted").Inc()
	s.Fire(dto.WEBHOOK_DELETED, w)
	WebsocketService().Broadcast(WEBHOOK_DELETED, w.ToDto())

	return nil
}

func (s *webhookSvc) GetWebhookById(uuid string) (*model.Webhook, error) {
	return s.webhookRepository.First(uuid)
}

func (s *webhookSvc) UpdateWebhook(webhookUuid string, webhook *dto.NewWebhook) (*model.Webhook, error) {
	w, err := s.GetWebhookById(webhookUuid)
	if err != nil {
		return nil, err
	}

	w.Event = webhook.Event
	w.Url = webhook.Url

	w, err = s.webhookRepository.Update(w)
	if err != nil {
		return nil, err
	}

	s.sev.Metrics().Gauge("webhook.updated").Inc()
	s.Fire(dto.WEBHOOK_UPDATED, w)
	WebsocketService().Broadcast(WEBHOOK_UPDATED, w.ToDto())

	return w, nil
}

func (s *webhookSvc) NewWebhook(webhook *dto.NewWebhook) (*model.Webhook, error) {
	w, err := s.webhookRepository.Create(webhook.Event, webhook.Url)
	s.sev.Logger().Infof("created new webhook for event %s (uuid: %s)", w.Event, w.Uuid)

	s.sev.Metrics().Gauge("webhook.created").Inc()
	s.Fire(dto.WEBHOOK_CREATED, w)
	WebsocketService().Broadcast(WEBHOOK_CREATED, w.ToDto())

	return w, err
}

func (s *webhookSvc) Fire(event dto.WebhookEvent, data interface{}) error {
	webhooks, err := s.webhookRepository.ListByEvent(event)
	for _, webhook := range *webhooks {
		go s.sev.FireWebhook(&webhook, data)
		s.sev.Metrics().Gauge("webhook.executed").Inc()
	}
	return err
}
