package repository

import (
	"errors"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"goyave.dev/goyave/v5/database"
)

type Client struct {
	DB *gorm.DB
}

func (r *Client) Setup() *Client {
	_ = r.DB.AutoMigrate(&model.Client{})
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
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Labels").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "identifier"}},
			UpdateAll: true,
		}).Create(newClient).Error; err != nil {
			return err
		}

		if err := upsertLabels(tx, newClient.Labels); err != nil {
			return err
		}

		return tx.Model(newClient).Association("Labels").Replace(newClient.Labels)
	})
	if err != nil {
		return nil, err
	}

	return newClient, nil
}

func (r *Client) Self(identifier string) (*model.Client, error) {
	var client model.Client
	result := r.DB.Preload("Labels").Where("identifier = ?", identifier).First(&client)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &client, nil
}

func (r *Client) First() (*model.Client, error) {
	var client model.Client
	result := r.DB.Preload("Labels").First(&client)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &client, nil
}
