package repository

import (
	"errors"

	"github.com/welovemedia/ffmate/internal/database/model"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
)

type Client struct {
	DB *gorm.DB
}

func (r *Client) Setup() *Client {
	r.DB.AutoMigrate(&model.Client{})
	return r
}

func (r *Client) List(page int, perPage int) (*[]model.Client, int64, error) {
	var tasks = &[]model.Client{}
	tx := r.DB.Preload("Labels").Order("last_seen DESC")
	d := database.NewPaginator(tx, page+1, perPage, tasks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Client) Save(newClient *model.Client) (*model.Client, error) {
	if err := r.DB.Save(newClient).Error; err != nil {
		return nil, err
	}

	for i := range newClient.Labels {
		r.DB.FirstOrCreate(&newClient.Labels[i], model.Label{Value: newClient.Labels[i].Value})
	}

	r.DB.Model(newClient).Association("Labels").Replace(newClient.Labels)

	return newClient, nil
}

func (m *Client) Self(identifier string) (*model.Client, error) {
	var client model.Client
	result := m.DB.Preload("Labels").Where("identifier = ?", identifier).First(&client)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &client, nil
}

func (m *Client) First() (*model.Client, error) {
	var client model.Client
	result := m.DB.Preload("Labels").First(&client)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &client, nil
}

func (r *Client) Count() (int64, error) {
	var count int64
	db := r.DB.Model(&model.Client{}).Count(&count)
	return count, db.Error
}
