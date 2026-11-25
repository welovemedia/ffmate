package dto

import "time"

type APIKeyAccess string

const (
	APIKeyAccessRead  APIKeyAccess = "read"
	APIKeyAccessWrite APIKeyAccess = "write"
)

type NewAPIKey struct {
	Name   string       `json:"name"`
	Access APIKeyAccess `json:"access"`
}

type APIKey struct {
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	UUID      string       `json:"uuid"`
	Name      string       `json:"name"`
	Access    APIKeyAccess `json:"access"`
	Prefix    string       `json:"prefix"`
}

// CreatedAPIKey is returned once on creation (and rotation) and is the only time the plain key is exposed
type CreatedAPIKey struct {
	APIKey
	Key string `json:"key"`
}
