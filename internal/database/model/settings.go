package model

import (
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

type Settings struct {
	ID uint `gorm:"primaryKey;unique"`
}

func (r *Settings) ToDTO() *dto.Settings {
	return &dto.Settings{}
}

func (Settings) TableName() string {
	return "settings"
}
