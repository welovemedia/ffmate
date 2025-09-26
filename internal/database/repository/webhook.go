package repository

import (
	"errors"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
)

type Webhook struct {
	DB *gorm.DB
}

func (r *Webhook) Setup() *Webhook {
	_ = r.DB.AutoMigrate(&model.Webhook{})
	return r
}

func (r *Webhook) First(uuid string) (*model.Webhook, error) {
	var webhook model.Webhook
	result := r.DB.Where("uuid = ?", uuid).First(&webhook)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &webhook, nil
}

func (r *Webhook) Update(webhook *model.Webhook) (*model.Webhook, error) {
	db := r.DB.Save(webhook)
	return webhook, db.Error
}

func (r *Webhook) Delete(w *model.Webhook) error {
	r.DB.Delete(w)
	return r.DB.Error
}

func (r *Webhook) Add(newWebhook *model.Webhook) (*model.Webhook, error) {
	db := r.DB.Create(newWebhook)
	return newWebhook, db.Error
}

func (r *Webhook) List(page int, perPage int) (*[]model.Webhook, int64, error) {
	var webhooks = &[]model.Webhook{}
	tx := r.DB.Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, webhooks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Webhook) ListAllByEvent(event dto.WebhookEvent) (*[]model.Webhook, error) {
	var webhooks = &[]model.Webhook{}
	r.DB.Order("created_at DESC").Where("event = ?", event).Find(&webhooks)
	return webhooks, r.DB.Error
}

func (r *Webhook) Count() (int64, error) {
	var count int64
	db := r.DB.Model(&model.Webhook{}).Count(&count)
	return count, db.Error
}

/**
 * Stats (telemetry) related methods
 */

func (r *Webhook) CountDeleted() (int64, error) {
	var count int64
	result := r.DB.Unscoped().Model(&model.Webhook{}).Unscoped().Where("deleted_at IS NOT NULL").Count(&count)
	return count, result.Error
}
