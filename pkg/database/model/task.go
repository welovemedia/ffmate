package model

import (
	"time"

	"github.com/welovemedia/ffmate/pkg/dto"
	"gorm.io/gorm"
)

type Task struct {
	ID uint `gorm:"primarykey"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Uuid string

	Command    string
	InputFile  string
	OutputFile string

	Status   dto.TaskStatus
	Progress uint
}

func (m *Task) ToDto() *dto.Task {
	return &dto.Task{
		Uuid: m.Uuid,

		Command:    m.Command,
		InputFile:  m.InputFile,
		OutputFile: m.OutputFile,

		Status:   m.Status,
		Progress: m.Progress,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (Task) TableName() string {
	return "tasks"
}
