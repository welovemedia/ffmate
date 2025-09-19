package model

import (
	"time"

	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
)

type WebhookExecution struct {
	ID uint `gorm:"primarykey"`

	Uuid string

	Event dto.WebhookEvent
	Url   string

	Request  *dto.WebhookRequest  `gorm:"json"`
	Response *dto.WebhookResponse `gorm:"json"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *WebhookExecution) ToDto() *dto.WebhookExecution {
	return &dto.WebhookExecution{
		Uuid: m.Uuid,

		Event: m.Event,
		Url:   m.Url,

		Request:  m.Request,
		Response: m.Response,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (WebhookExecution) TableName() string {
	return "webhookExecution"
}
