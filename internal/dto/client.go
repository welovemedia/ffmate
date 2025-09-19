package dto

type Client struct {
	Identifier string `json:"identifier"`
	Session    string `json:"session"`
	Cluster    string `json:"cluster"`

	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Version string `json:"version"`
	FFMpeg  string `json:"ffmpeg"`

	LastSeen int64 `json:"lastSeen"`

	Self bool `json:"self,omitempty"`
}

type NewClient struct {
	Identifier string `json:"identifier"`
	Session    string `json:"session"`
	Cluster    string `json:"cluster"`
}
