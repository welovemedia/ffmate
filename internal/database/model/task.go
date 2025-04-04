package model

import (
	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
)

type Task struct {
	ID uint `gorm:"primarykey"`

	CreatedAt int64          `gorm:"autoCreateTime:milli"`
	UpdatedAt int64          `gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Uuid  string
	Batch string

	Name string

	Command    *dto.RawResolved `gorm:"type:json"`
	InputFile  *dto.RawResolved `gorm:"type:json"`
	OutputFile *dto.RawResolved `gorm:"type:json"`

	Metadata *dto.InterfaceMap `gorm:"serializer:json"` // Additional metadata for the task

	Status    dto.TaskStatus `gorm:"index"`
	Error     string
	Progress  float64
	Remaining float64

	Priority uint

	PreProcessing  *dto.PrePostProcessing `gorm:"type:json"`
	PostProcessing *dto.PrePostProcessing `gorm:"type:json"`

	Source string

	Session string

	StartedAt  int64
	FinishedAt int64
}

func (m *Task) ToDto() *dto.Task {
	return &dto.Task{
		Uuid: m.Uuid,

		Name:  m.Name,
		Batch: m.Batch,

		Command:    m.Command,
		InputFile:  m.InputFile,
		OutputFile: m.OutputFile,

		Metadata: m.Metadata,

		Status:    m.Status,
		Progress:  m.Progress,
		Remaining: m.Remaining,

		Error: m.Error,

		Source: m.Source,

		Priority: m.Priority,

		PreProcessing:  m.PreProcessing,
		PostProcessing: m.PostProcessing,

		StartedAt:  m.StartedAt,
		FinishedAt: m.FinishedAt,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (Task) TableName() string {
	return "tasks"
}
