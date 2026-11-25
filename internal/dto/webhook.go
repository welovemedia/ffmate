package dto

import (
	"time"
)

type WebhookEvent string

const (
	BatchCreated  WebhookEvent = "batch.created"
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

// TaskWebhookEvents are the events fired for webhooks defined directly on a task
var TaskWebhookEvents = []WebhookEvent{TaskCreated, TaskUpdated, TaskDeleted}

// PresetWebhookEvents are the events fired for webhooks defined on a preset
// (preset.* for the preset itself, task.* are passed on to tasks created from the preset)
var PresetWebhookEvents = append([]WebhookEvent{PresetCreated, PresetUpdated, PresetDeleted}, TaskWebhookEvents...)

// WebhookEvents are all events that can be subscribed to with a global webhook
var WebhookEvents = []WebhookEvent{
	BatchCreated, BatchFinished,
	TaskCreated, TaskUpdated, TaskDeleted,
	PresetCreated, PresetUpdated, PresetDeleted,
	WebhookCreated, WebhookUpdated, WebhookDeleted,
	WatchfolderCreated, WatchfolderUpdated, WatchfolderDeleted,
}

// WebhookEventNames returns the given events as strings (e.g. for validation rules)
func WebhookEventNames(events []WebhookEvent) []string {
	names := make([]string, len(events))
	for i, e := range events {
		names[i] = string(e)
	}
	return names
}

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
