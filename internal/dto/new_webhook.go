package dto

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
