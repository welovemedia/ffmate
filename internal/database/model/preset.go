package model

import (
	"time"

	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
)

type Preset struct {
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Webhooks       *dto.DirectWebhooks       `gorm:"type:jsonb"`
	PostProcessing *dto.NewPrePostProcessing `gorm:"type:jsonb"`
	PreProcessing  *dto.NewPrePostProcessing `gorm:"type:jsonb"`
	DeletedAt      gorm.DeletedAt            `gorm:"index"`
	Name           string
	OutputFile     string
	Command        string
	Labels         []Label `gorm:"many2many:preset_labels;"`
	UUID           string
	Description    string
	Retries        int
	Priority       uint
	ID             uint `gorm:"primarykey"`
}

func (m *Preset) ToDTO() *dto.Preset {
	var labels = make([]string, len(m.Labels))
	for i, label := range m.Labels {
		labels[i] = label.Value
	}

	return &dto.Preset{
		UUID: m.UUID,

		Command:     m.Command,
		Name:        m.Name,
		Description: m.Description,

		OutputFile: m.OutputFile,

		Priority: m.Priority,

		Webhooks: m.Webhooks,

		Labels: labels,
		Retries: m.Retries,

		PreProcessing:  m.PreProcessing,
		PostProcessing: m.PostProcessing,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (Preset) TableName() string {
	return "presets"
}
