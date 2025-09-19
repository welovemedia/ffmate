package dto

type RawResolved struct {
	Raw      string `json:"raw"`
	Resolved string `json:"resolved,omitempty"`
}
