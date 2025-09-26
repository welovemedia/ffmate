package repository

import (
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"gorm.io/gorm"
)

type Settings struct {
	DB *gorm.DB
}

func (r *Settings) Setup() *Settings {
	_ = r.DB.AutoMigrate(&model.Settings{})
	return r
}

func (r *Settings) Load() (*model.Settings, error) {
	settings := &model.Settings{}
	db := r.DB.First(&settings, 1)
	if db.Error != nil && db.Error == gorm.ErrRecordNotFound {
		settings = &model.Settings{}
		db.Error = nil
	}
	return settings, db.Error
}

func (r *Settings) Store(newSetting *model.Settings) (*model.Settings, error) {
	newSetting.ID = 1 // enforce single row
	result := r.DB.Save(newSetting)
	return newSetting, result.Error
}
