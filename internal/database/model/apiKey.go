package model

import (
	"time"

	"github.com/welovemedia/ffmate/v2/internal/dto"
)

type APIKey struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	UUID      string `gorm:"uniqueIndex"`
	Name      string
	Access    dto.APIKeyAccess
	Prefix    string
	KeyHash   string `gorm:"uniqueIndex"` // sha256 (hex) of the plain key, the plain key is never stored
	ID        uint   `gorm:"primarykey"`
}

func (m *APIKey) ToDTO() *dto.APIKey {
	return &dto.APIKey{
		UUID:   m.UUID,
		Name:   m.Name,
		Access: m.Access,
		Prefix: m.Prefix,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (APIKey) TableName() string {
	return "api_keys"
}
