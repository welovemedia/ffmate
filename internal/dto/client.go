package dto

type Client struct {
	Identifier string `json:"identifier"`
	Session    string `json:"session"`
	Cluster    string `json:"cluster"`
	Labels     Labels `json:"labels,omitempty"`

	OS                 string `json:"os"`
	Arch               string `json:"arch"`
	Version            string `json:"version"`
	FFMpeg             string `json:"ffmpeg"`
	MaxConcurrentTasks int    `json:"maxConcurrentTasks"`

	LastSeen int64 `json:"lastSeen"`

	Self bool `json:"self,omitempty"`
}
