package dto

type NewPrePostProcessing struct {
	ScriptPath    string `json:"scriptPath,omitempty"`
	SidecarPath   string `json:"sidecarPath,omitempty"`
	ImportSidecar bool   `json:"importSidecar,omitempty"`
}

type PrePostProcessing struct {
	ScriptPath    *RawResolved `json:"scriptPath,omitempty"`
	SidecarPath   *RawResolved `json:"sidecarPath,omitempty"`
	Error         string       `json:"error,omitempty"`
	StartedAt     int64        `json:"startedAt,omitempty"`
	FinishedAt    int64        `json:"finishedAt,omitempty"`
	ImportSidecar bool         `json:"importSidecar,omitempty"`
}
