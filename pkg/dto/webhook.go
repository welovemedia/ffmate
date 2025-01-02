package dto

import "time"

type Webhook struct {
	Event WebhookEvent `json:"event"`
	Url   string       `json:"url"`

	Uuid string `json:"uuid"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
