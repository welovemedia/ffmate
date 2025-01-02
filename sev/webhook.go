package sev

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/dto"
)

type eventMessage struct {
	Event dto.WebhookEvent `json:"event"`
	Data  interface{}      `json:"data"`
}

func (s *Sev) FireWebhook(webhook *model.Webhook, data interface{}) {
	msg := eventMessage{
		Event: webhook.Event,
		Data:  data,
	}
	b, err := json.Marshal(&msg)
	if err != nil {
		s.Logger().Warnf("failed to fire webhook for event '%s' (uuid: %s) due to marshalling problems: %+v", webhook.Event, webhook.Uuid, err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", webhook.Url, bytes.NewBuffer(b))
	if err != nil {
		s.Logger().Errorf("failed to create http request", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", s.appName+"/"+s.appVersion)

	_, err = client.Do(req)
	if err != nil {
		s.Logger().Warnf("failed to fire webhook for event '%s' (uuid: %s) due to http problems: %+v", webhook.Event, webhook.Uuid, err)
	} else {
		s.Logger().Debugf("fired webhook for event '%s' (uuid: %s)", webhook.Event, webhook.Uuid)
	}
}
