package model

import (
	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
)

type Watchfolder struct {
	ID uint `gorm:"primarykey"`

	CreatedAt int64          `gorm:"autoCreateTime:milli"`
	UpdatedAt int64          `gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Uuid string

	Name        string
	Description string

	Path         string
	Interval     int
	GrowthChecks int

	Filter *dto.WatchfolderFilter

	Preset string
	Labels []Label `gorm:"many2many:watchfolder_labels;"`

	Suspended bool

	LastRun int64

	Error     string
	LastCheck int64
}

func (m *Watchfolder) ToDto() *dto.Watchfolder {
	var labels = make([]string, len(m.Labels))
	for i, label := range m.Labels {
		labels[i] = label.Value
	}

	return &dto.Watchfolder{
		Uuid: m.Uuid,

		Name:        m.Name,
		Description: m.Description,

		Path:         m.Path,
		Interval:     m.Interval,
		GrowthChecks: m.GrowthChecks,

		Preset: m.Preset,
		Labels: labels,

		Filter: m.Filter,

		Suspended: m.Suspended,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,

		Error:     m.Error,
		LastCheck: m.LastCheck,
	}
}

func (Watchfolder) TableName() string {
	return "watchfolder"
}
