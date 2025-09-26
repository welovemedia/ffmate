package model

import (
	"github.com/welovemedia/ffmate/v2/internal/dto"
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

	Command    *dto.RawResolved `gorm:"type:jsonb"`
	InputFile  *dto.RawResolved `gorm:"type:jsonb"`
	OutputFile *dto.RawResolved `gorm:"type:jsonb"`

	Metadata *dto.MetadataMap `gorm:"serializer:json"` // Additional metadata for the task

	Status    dto.TaskStatus `gorm:"index"`
	Error     string
	Progress  float64
	Remaining float64

	Priority uint

	Webhooks *dto.DirectWebhooks `gorm:"type:jsonb"`

	PreProcessing  *dto.PrePostProcessing `gorm:"type:jsonb"`
	PostProcessing *dto.PrePostProcessing `gorm:"type:jsonb"`

	Source dto.TaskSource

	ClientIdentifier string  `gorm:"index"`
	Client           *Client `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ClientIdentifier;references:Identifier"`

	StartedAt  int64
	FinishedAt int64
}

func (m *Task) ToDto() *dto.Task {
	d := &dto.Task{
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

		Webhooks: m.Webhooks,

		PreProcessing:  m.PreProcessing,
		PostProcessing: m.PostProcessing,

		StartedAt:  m.StartedAt,
		FinishedAt: m.FinishedAt,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.Client != nil {
		d.Client = &dto.Client{
			Identifier: m.Client.Identifier,
			Session:    m.Client.Session,
			Cluster:    m.Client.Cluster,
			OS:         m.Client.OS,
			Arch:       m.Client.Arch,
			Version:    m.Client.Version,
			LastSeen:   m.Client.LastSeen,
		}
	}

	return d
}

func (Task) TableName() string {
	return "tasks"
}
