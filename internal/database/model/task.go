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

	Resolved *dto.Resolved `gorm:"type:json"`

	Command    string
	InputFile  string
	OutputFile string

	Status   dto.TaskStatus
	Error    string
	Progress float64

	Priority uint

	PostProcessing *dto.PostProcessing `gorm:"type:json"`

	StartedAt  int64
	FinishedAt int64
}

func (m *Task) ToDto() *dto.Task {
	return &dto.Task{
		Uuid: m.Uuid,

		Name:  m.Name,
		Batch: m.Batch,

		Resolved: m.Resolved,

		Command:    m.Command,
		InputFile:  m.InputFile,
		OutputFile: m.OutputFile,

		Status:   m.Status,
		Progress: m.Progress,
		Error:    m.Error,

		Priority: m.Priority,

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
