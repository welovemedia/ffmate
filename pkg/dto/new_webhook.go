package dto

type WebhookEvent string

const (
	TASK_CREATED        WebhookEvent = "task.created"
	TASK_STATUS_UPDATED WebhookEvent = "task.status.updated"

	WEBHOOK_CREATED WebhookEvent = "webhook.created"
	WEBHOOK_DELETED WebhookEvent = "webhook.deleted"
)

type NewWebhook struct {
	Event WebhookEvent `json:"event"`
	Url   string       `json:"url"`
}
