package model

import (
	"time"

	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
)

type WebhookExecution struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Request   *dto.WebhookRequest  `gorm:"json"`
	Response  *dto.WebhookResponse `gorm:"json"`
	DeletedAt gorm.DeletedAt       `gorm:"index"`
	UUID      string
	Event     dto.WebhookEvent
	URL       string
	ID        uint `gorm:"primarykey"`
}

func (m *WebhookExecution) ToDTO() *dto.WebhookExecution {
	return &dto.WebhookExecution{
		UUID: m.UUID,

		Event: m.Event,
		URL:   m.URL,

		Request:  m.Request,
		Response: m.Response,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (WebhookExecution) TableName() string {
	return "webhookExecution"
}
