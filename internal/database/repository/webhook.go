package repository

import (
	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
)

type Webhook struct {
	DB *gorm.DB
}

func (t *Webhook) Setup() {
	t.DB.AutoMigrate(&model.Webhook{})
}

func (m *Webhook) List(page int, perPage int) (*[]model.Webhook, int64, error) {
	total, _ := m.Count()
	var webhooks = &[]model.Webhook{}
	m.DB.Order("event ASC, created_at DESC").Limit(perPage).Offset(page * perPage).Find(&webhooks)
	return webhooks, total, m.DB.Error
}

func (m *Webhook) First(uuid string) (*model.Webhook, error) {
	var webhook = &model.Webhook{}
	err := m.DB.Where("uuid = ?", uuid).First(&webhook).Error
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

func (m *Webhook) Count() (int64, error) {
	var count int64
	db := m.DB.Model(&model.Webhook{}).Count(&count)
	return count, db.Error
}

func (m *Webhook) CountDeleted() (int64, error) {
	var count int64
	db := m.DB.Unscoped().Model(&model.Webhook{}).Where("deleted_at IS NOT NULL").Count(&count)
	return count, db.Error
}

func (m *Webhook) Update(w *model.Webhook) (*model.Webhook, error) {
	m.DB.Save(w)
	return w, m.DB.Error
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
