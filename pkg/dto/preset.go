package dto

import "time"

type Preset struct {
	Uuid string `json:"uuid"`

	Command     string `json:"command"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
