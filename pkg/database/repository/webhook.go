package repository

import (
	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/dto"
	"gorm.io/gorm"
)

type Webhook struct {
	DB *gorm.DB
}

func (t *Webhook) Setup() {
	t.DB.AutoMigrate(&model.Webhook{})
}

func (m *Webhook) List() (*[]model.Webhook, error) {
	var webhooks = &[]model.Webhook{}
	m.DB.Order("event ASC, created_at DESC").Find(&webhooks)
	return webhooks, m.DB.Error
}

func (m *Webhook) First(uuid string) (*model.Webhook, error) {
	var webhook = &model.Webhook{}
	m.DB.Where("uuid = ?", uuid).Find(&webhook)
	return webhook, m.DB.Error
}

func (m *Webhook) Count() (int64, error) {
	var count int64
	db := m.DB.Model(&model.Webhook{}).Count(&count)
	return count, db.Error
}

func (m *Webhook) Delete(w *model.Webhook) error {
	m.DB.Delete(w)
	return m.DB.Error
}

func (m *Webhook) ListByEvent(event dto.WebhookEvent) (*[]model.Webhook, error) {
	var webhooks = &[]model.Webhook{}
	m.DB.Order("created_at DESC").Where("event = ?", event).Find(&webhooks)
	return webhooks, m.DB.Error
}

func (m *Webhook) Create(event dto.WebhookEvent, url string) (*model.Webhook, error) {
	webhook := &model.Webhook{Uuid: uuid.NewString(), Event: event, Url: url}
	db := m.DB.Create(webhook)
	return webhook, db.Error
}
