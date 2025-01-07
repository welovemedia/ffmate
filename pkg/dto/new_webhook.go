package dto

type WebhookEvent string

const (
	TASK_CREATED        WebhookEvent = "task.created"
	TASK_STATUS_UPDATED WebhookEvent = "task.status.updated"

	PRESET_CREATED WebhookEvent = "preset.created"
	PRESET_DELETE  WebhookEvent = "preset.deleted"

	WEBHOOK_CREATED WebhookEvent = "webhook.created"
	WEBHOOK_DELETED WebhookEvent = "webhook.deleted"
)

type NewWebhook struct {
	Event WebhookEvent `json:"event"`
	Url   string       `json:"url"`
}
