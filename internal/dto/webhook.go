package dto

import (
	"time"
)

type WebhookEvent string

const (
	BatCreated    WebhookEvent = "batch.created"
	BatchFinished WebhookEvent = "batch.finished"

	TaskCreated WebhookEvent = "task.created"
	TaskUpdated WebhookEvent = "task.updated"
	TaskDeleted WebhookEvent = "task.deleted"

	PresetCreated WebhookEvent = "preset.created"
	PresetUpdated WebhookEvent = "preset.updated"
	PresetDeleted WebhookEvent = "preset.deleted"

	WebhookCreated WebhookEvent = "webhook.created"
	WebhookUpdated WebhookEvent = "webhook.updated"
	WebhookDeleted WebhookEvent = "webhook.deleted"

	WatchfolderCreated WebhookEvent = "watchfolder.created"
	WatchfolderUpdated WebhookEvent = "watchfolder.updated"
	WatchfolderDeleted WebhookEvent = "watchfolder.deleted"
)

type NewWebhook struct {
	Event WebhookEvent `json:"event"`
	URL   string       `json:"url"`
}

type Webhook struct {
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Event     WebhookEvent `json:"event"`
	URL       string       `json:"url"`
	UUID      string       `json:"uuid"`
}

type WebhookExecution struct {
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	Request   *WebhookRequest  `json:"request"`
	Response  *WebhookResponse `json:"response"`
	UUID      string           `json:"uuid"`
	Event     WebhookEvent     `json:"event"`
	URL       string           `json:"url"`
}

type WebhookRequest struct {
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type WebhookResponse struct {
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
	Status  int                 `json:"status"`
}

type DirectWebhooks []NewWebhook
