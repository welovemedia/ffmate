package dto

import (
	"time"
)

type WebhookEvent string

const (
	BATCH_CREATED  WebhookEvent = "batch.created"
	BATCH_FINISHED WebhookEvent = "batch.finished"

	TASK_CREATED WebhookEvent = "task.created"
	TASK_UPDATED WebhookEvent = "task.updated"
	TASK_DELETED WebhookEvent = "task.deleted"

	PRESET_CREATED WebhookEvent = "preset.created"
	PRESET_UPDATED WebhookEvent = "preset.updated"
	PRESET_DELETED WebhookEvent = "preset.deleted"

	WEBHOOK_CREATED WebhookEvent = "webhook.created"
	WEBHOOK_UPDATED WebhookEvent = "webhook.updated"
	WEBHOOK_DELETED WebhookEvent = "webhook.deleted"

	WATCHFOLDER_CREATED WebhookEvent = "watchfolder.created"
	WATCHFOLDER_UPDATED WebhookEvent = "watchfolder.updated"
	WATCHFOLDER_DELETED WebhookEvent = "watchfolder.deleted"
)

type NewWebhook struct {
	Event WebhookEvent `json:"event"`
	Url   string       `json:"url"`
}

type Webhook struct {
	Event WebhookEvent `json:"event"`
	Url   string       `json:"url"`

	Uuid string `json:"uuid"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WebhookExecution struct {
	Uuid string `json:"uuid"`

	Event WebhookEvent `json:"event"`
	Url   string       `json:"url"`

	Request  *WebhookRequest  `json:"request"`
	Response *WebhookResponse `json:"response"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
