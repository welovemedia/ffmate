package model

import (
	"time"

	"github.com/welovemedia/ffmate/v2/internal/dto"
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

	OutputFile string

	Priority uint

	Labels []Label `gorm:"many2many:preset_labels;"`

	Webhooks *dto.DirectWebhooks `gorm:"type:jsonb"`

	PreProcessing  *dto.NewPrePostProcessing `gorm:"type:jsonb"`
	PostProcessing *dto.NewPrePostProcessing `gorm:"type:jsonb"`

	Description string
}

func (m *Preset) ToDto() *dto.Preset {
	var labels = make([]string, len(m.Labels))
	for i, label := range m.Labels {
		labels[i] = label.Value
	}

	return &dto.Preset{
		Uuid: m.Uuid,

		Command:     m.Command,
		Name:        m.Name,
		Description: m.Description,

		OutputFile: m.OutputFile,

		Priority: m.Priority,

		Webhooks: m.Webhooks,

		Labels: labels,

		PreProcessing:  m.PreProcessing,
		PostProcessing: m.PostProcessing,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (Preset) TableName() string {
	return "presets"
}
