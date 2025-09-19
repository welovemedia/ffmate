package dto

import "time"

type NewPreset struct {
	Command string `json:"command"`

	Priority uint `json:"priority"`

	OutputFile string `json:"outputFile"`

	Webhooks *DirectWebhooks `json:"webhooks"`

	PreProcessing  *NewPrePostProcessing `json:"preProcessing"`
	PostProcessing *NewPrePostProcessing `json:"postProcessing"`

	Name        string `json:"name"`
	Description string `json:"description"`

	GlobalPresetName string `json:"globalPresetName"`
}

type Preset struct {
	Uuid string `json:"uuid"`

	Command     string `json:"command"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	OutputFile string `json:"outputFile"`

	Priority uint `json:"priority"`

	PreProcessing  *NewPrePostProcessing `json:"preProcessing,omitempty"`
	PostProcessing *NewPrePostProcessing `json:"postProcessing,omitempty"`

	Webhooks *DirectWebhooks `json:"webhooks,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
