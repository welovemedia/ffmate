package model

import (
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

type Client struct {
	Identifier string `gorm:"primaryKey"`

	Session string
	Cluster string
	Labels  []Label `gorm:"many2many:client_labels;"`

	OS                 string
	Arch               string
	Version            string
	FFMpeg             string
	MaxConcurrentTasks int

	LastSeen int64
}

func (c *Client) ToDTO() *dto.Client {
	var labels = make([]string, len(c.Labels))
	for i, label := range c.Labels {
		labels[i] = label.Value
	}

	client := &dto.Client{
		Identifier:         c.Identifier,
		Session:            c.Session,
		Cluster:            c.Cluster,
		OS:                 c.OS,
		Arch:               c.Arch,
		Labels:             labels,
		Version:            c.Version,
		FFMpeg:             c.FFMpeg,
		MaxConcurrentTasks: c.MaxConcurrentTasks,
		LastSeen:           c.LastSeen,
	}

	if c.Session == cfg.GetString("ffmate.session") {
		client.Self = true
	}

	return client
}

func (Client) TableName() string {
	return "client"
}
