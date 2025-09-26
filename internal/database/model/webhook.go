package model

import (
	"time"

	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
)

type Webhook struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UUID      string
	Event     dto.WebhookEvent
	URL       string
	ID        uint `gorm:"primarykey"`
}

func (m *Webhook) ToDTO() *dto.Webhook {
	return &dto.Webhook{
		Event: m.Event,
		URL:   m.URL,

		UUID: m.UUID,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
