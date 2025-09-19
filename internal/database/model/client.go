package model

import (
	"github.com/welovemedia/ffmate/internal/cfg"
	"github.com/welovemedia/ffmate/internal/dto"
)

type Client struct {
	Identifier string `gorm:"primaryKey"`

	Session string
	Cluster string

	OS      string
	Arch    string
	Version string
	FFMpeg  string

	LastSeen int64
}

func (c *Client) ToDto() *dto.Client {
	client := &dto.Client{
		Identifier: c.Identifier,
		Session:    c.Session,
		Cluster:    c.Cluster,
		OS:         c.OS,
		Arch:       c.Arch,
		Version:    c.Version,
		FFMpeg:     c.FFMpeg,
		LastSeen:   c.LastSeen,
	}

	if c.Session == cfg.GetString("ffmate.session") {
		client.Self = true
	}

	return client
}

func (Client) TableName() string {
	return "client"
}
