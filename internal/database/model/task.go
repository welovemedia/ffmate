package model

import (
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
)

type Task struct {
	OutputFile       *dto.RawResolved       `gorm:"type:jsonb"`
	Client           *Client                `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ClientIdentifier;references:Identifier"`
	PostProcessing   *dto.PrePostProcessing `gorm:"type:jsonb"`
	PreProcessing    *dto.PrePostProcessing `gorm:"type:jsonb"`
	Webhooks         *dto.DirectWebhooks    `gorm:"type:jsonb"`
	Metadata         *dto.MetadataMap       `gorm:"serializer:json"`
	Command          *dto.RawResolved       `gorm:"type:jsonb"`
	InputFile        *dto.RawResolved       `gorm:"type:jsonb"`
	DeletedAt        gorm.DeletedAt         `gorm:"index"`
	Name             string
	Source           dto.TaskSource
	Status           dto.TaskStatus `gorm:"index"`
	Error            string
	ClientIdentifier string `gorm:"index"`
	UUID             string
	Batch            string
	Labels           []Label `gorm:"many2many:task_labels;"`
	Priority         uint
	Retries          int
	Retried          int
	Remaining        float64
	UpdatedAt        int64 `gorm:"autoUpdateTime:milli"`
	ID               uint  `gorm:"primarykey"`
	Progress         float64
	CreatedAt        int64 `gorm:"autoCreateTime:milli"`
	StartedAt        int64
	FinishedAt       int64
}

func (m *Task) ToDTO() *dto.Task {
	var labels = make([]string, len(m.Labels))
	for i, label := range m.Labels {
		labels[i] = label.Value
	}

	d := &dto.Task{
		UUID: m.UUID,

		Name:  m.Name,
		Batch: m.Batch,

		Command:    m.Command,
		InputFile:  m.InputFile,
		OutputFile: m.OutputFile,

		Metadata: m.Metadata,

		Status:    m.Status,
		Progress:  m.Progress,
		Remaining: m.Remaining,

		Labels: labels,
		Retries: m.Retries,
		Retried: m.Retried,

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
			Identifier:         m.Client.Identifier,
			Session:            m.Client.Session,
			Cluster:            m.Client.Cluster,
			OS:                 m.Client.OS,
			Arch:               m.Client.Arch,
			Version:            m.Client.Version,
			MaxConcurrentTasks: m.Client.MaxConcurrentTasks,
			LastSeen:           m.Client.LastSeen,
		}
	}

	return d
}

func (Task) TableName() string {
	return "tasks"
}
