package model

import (
	"time"

	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
)

type Webhook struct {
	ID uint `gorm:"primarykey"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Uuid string

	Event dto.WebhookEvent
	Url   string
}

func (m *Webhook) ToDto() *dto.Webhook {
	return &dto.Webhook{
		Event: m.Event,
		Url:   m.Url,

		Uuid: m.Uuid,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
