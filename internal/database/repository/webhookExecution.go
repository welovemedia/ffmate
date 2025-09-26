package repository

import (
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
)

type WebhookExecution struct {
	DB *gorm.DB
}

func (r *WebhookExecution) Setup() *WebhookExecution {
	r.DB.AutoMigrate(&model.WebhookExecution{})
	return r
}

func (r *WebhookExecution) Add(newWebhookExecution *model.WebhookExecution) (*model.WebhookExecution, error) {
	db := r.DB.Create(newWebhookExecution)
	return newWebhookExecution, db.Error
}

func (r *WebhookExecution) List(page int, perPage int) (*[]model.WebhookExecution, int64, error) {
	var webhookExecutions = &[]model.WebhookExecution{}
	tx := r.DB.Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, webhookExecutions)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *WebhookExecution) Count() (int64, error) {
	var count int64
	db := r.DB.Model(&model.WebhookExecution{}).Count(&count)
	return count, db.Error
}
