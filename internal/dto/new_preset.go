package dto

type NewPreset struct {
	Command     string `json:"command"`
	Priority    uint   `json:"priority"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
