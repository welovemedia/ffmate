package model

import (
	"time"

	"github.com/welovemedia/ffmate/pkg/dto"
	"gorm.io/gorm"
)

type Preset struct {
	ID uint `gorm:"primarykey"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Uuid string

	Command string
	Name    string
}

func (m *Preset) ToDto() *dto.Preset {
	return &dto.Preset{
		Uuid: m.Uuid,

		Command: m.Command,
		Name:    m.Name,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (Preset) TableName() string {
	return "presets"
}
