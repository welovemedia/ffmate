package repository

import (
	"errors"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
)

type APIKey struct {
	DB *gorm.DB
}

func (r *APIKey) Setup() *APIKey {
	_ = r.DB.AutoMigrate(&model.APIKey{})
	return r
}

func (r *APIKey) First(uuid string) (*model.APIKey, error) {
	return r.firstWhere("uuid = ?", uuid)
}

func (r *APIKey) FirstByHash(hash string) (*model.APIKey, error) {
	return r.firstWhere("key_hash = ?", hash)
}

func (r *APIKey) firstWhere(query string, arg any) (*model.APIKey, error) {
	var apiKey model.APIKey
	result := r.DB.Where(query, arg).First(&apiKey)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &apiKey, nil
}

func (r *APIKey) Add(apiKey *model.APIKey) (*model.APIKey, error) {
	db := r.DB.Create(apiKey)
	return apiKey, db.Error
}

func (r *APIKey) Save(apiKey *model.APIKey) (*model.APIKey, error) {
	db := r.DB.Save(apiKey)
	return apiKey, db.Error
}

func (r *APIKey) Delete(apiKey *model.APIKey) error {
	return r.DB.Delete(apiKey).Error
}

func (r *APIKey) List(page int, perPage int) (*[]model.APIKey, int64, error) {
	var apiKeys = &[]model.APIKey{}
	tx := r.DB.Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, apiKeys)
	err := d.Find()
	return d.Records, d.Total, err
}
