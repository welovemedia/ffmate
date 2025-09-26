package model

import (
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
)

type Watchfolder struct {
	Filter       *dto.WatchfolderFilter
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Error        string
	Preset       string
	UUID         string
	Name         string
	Description  string
	Path         string
	Interval     int
	GrowthChecks int
	ID           uint  `gorm:"primarykey"`
	UpdatedAt    int64 `gorm:"autoUpdateTime:milli"`
	LastRun      int64
	CreatedAt    int64 `gorm:"autoCreateTime:milli"`
	LastCheck    int64
	Suspended    bool
}

func (m *Watchfolder) ToDTO() *dto.Watchfolder {
	return &dto.Watchfolder{
		UUID: m.UUID,

		Name:        m.Name,
		Description: m.Description,

		Path:         m.Path,
		Interval:     m.Interval,
		GrowthChecks: m.GrowthChecks,

		Preset: m.Preset,

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
