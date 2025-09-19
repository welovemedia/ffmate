package model

import (
	"github.com/welovemedia/ffmate/internal/dto"
)

type Settings struct {
	ID uint `gorm:"primaryKey;unique"`
}

func (r *Settings) ToDto() *dto.Settings {
	return &dto.Settings{}
}

func (Settings) TableName() string {
	return "settings"
}
