package service

import (
	"errors"

	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/sev"
)

type WebhookService struct {
	Sev               *sev.Sev
	WebhookRepository *repository.Webhook
}

func (s *WebhookService) ListWebhooks() (*[]model.Webhook, error) {
	return s.WebhookRepository.List()
}

func (s *WebhookService) DeleteWebhook(uuid string) error {
	w, err := s.WebhookRepository.First(uuid)
	if err != nil {
		return err
	}

	if w.Uuid == "" {
		return errors.New("webhook for given uuid not found")
	}

	err = s.WebhookRepository.Delete(w)
	if err != nil {
		s.Sev.Logger().Warnf("failed to delete webhook for event %s (uuid: %s): %+v", w.Event, w.Uuid, err)
		return err
	}

	s.Sev.Logger().Infof("deleted webhook for event %s (uuid: %s)", w.Event, w.Uuid)

	s.Sev.Metrics().Gauge("webhook.deleted").Inc()
	s.Fire(dto.WEBHOOK_DELETED, w)

	return nil
}

func (s *WebhookService) NewWebhook(webhook *dto.NewWebhook) (*model.Webhook, error) {
	w, err := s.WebhookRepository.Create(webhook.Event, webhook.Url)
	s.Sev.Logger().Infof("created new webhook for event %s (uuid: %s)", w.Event, w.Uuid)

	s.Sev.Metrics().Gauge("webhook.created").Inc()
	s.Fire(dto.WEBHOOK_CREATED, w)

	return w, err
}

func (s *WebhookService) Fire(event dto.WebhookEvent, data interface{}) error {
	webhooks, err := s.WebhookRepository.ListByEvent(event)
	for _, webhook := range *webhooks {
		go s.Sev.FireWebhook(&webhook, data)
		s.Sev.Metrics().Gauge("webhook.executed").Inc()
	}
	return err
}
