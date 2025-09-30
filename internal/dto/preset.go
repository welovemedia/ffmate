package dto

import "time"

type NewPreset struct {
	Webhooks         *DirectWebhooks       `json:"webhooks"`
	PreProcessing    *NewPrePostProcessing `json:"preProcessing"`
	PostProcessing   *NewPrePostProcessing `json:"postProcessing"`
	Command          string                `json:"command"`
	OutputFile       string                `json:"outputFile"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Labels           Labels                `json:"labels"`
	Retries          int                   `json:"retries"`
	GlobalPresetName string                `json:"globalPresetName"`
	Priority         uint                  `json:"priority"`
}

type Preset struct {
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
	PreProcessing  *NewPrePostProcessing `json:"preProcessing,omitempty"`
	PostProcessing *NewPrePostProcessing `json:"postProcessing,omitempty"`
	Webhooks       *DirectWebhooks       `json:"webhooks,omitempty"`
	UUID           string                `json:"uuid"`
	Command        string                `json:"command"`
	Name           string                `json:"name"`
	Description    string                `json:"description,omitempty"`
	Retries        int                   `json:"retries"`
	OutputFile     string                `json:"outputFile"`
	Priority       uint                  `json:"priority"`
	Labels         Labels                `json:"labels,omitempty"`
}
