package dto

import "time"

type Preset struct {
	Uuid string `json:"uuid"`

	Name    string `json:"name"`
	Command string `json:"command"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
