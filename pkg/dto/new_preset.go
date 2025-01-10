package dto

type NewPreset struct {
	Command     string `json:"command"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
