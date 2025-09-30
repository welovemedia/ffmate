package dto

type TaskSource string

const (
	API         TaskSource = "api"
	WATCHFOLDER TaskSource = "watchfolder"
)

type TaskStatus string

const (
	Queued         TaskStatus = "QUEUED"
	Running        TaskStatus = "RUNNING"
	PreProcessing  TaskStatus = "PRE_PROCESSING"
	PostProcessing TaskStatus = "POST_PROCESSING"
	DoneSuccessful TaskStatus = "DONE_SUCCESSFUL"
	DoneError      TaskStatus = "DONE_ERROR"
	DoneCanceled   TaskStatus = "DONE_CANCELED"
)

type NewTask struct {
	Metadata       *MetadataMap          `json:"metadata"`
	Webhooks       *DirectWebhooks       `json:"webhooks"`
	PreProcessing  *NewPrePostProcessing `json:"preProcessing"`
	PostProcessing *NewPrePostProcessing `json:"postProcessing"`
	Command        string                `json:"command"`
	Preset         string                `json:"preset"`
	Name           string                `json:"name"`
	Labels         Labels                `json:"labels,omitempty"`
	Retries        int                   `json:"retries"`
	InputFile      string                `json:"inputFile"`
	OutputFile     string                `json:"outputFile"`
	Priority       uint                  `json:"priority"`
}

type Task struct {
	PostProcessing *PrePostProcessing `json:"postProcessing,omitempty"`
	Client         *Client            `json:"client,omitempty"`
	PreProcessing  *PrePostProcessing `json:"preProcessing,omitempty"`
	Command        *RawResolved       `json:"command"`
	InputFile      *RawResolved       `json:"inputFile"`
	OutputFile     *RawResolved       `json:"outputFile"`
	Metadata       *MetadataMap       `json:"metadata,omitempty"`
	Webhooks       *DirectWebhooks    `json:"webhooks,omitempty"`
	Status         TaskStatus         `json:"status"`
	Name           string             `json:"name,omitempty"`
	Batch          string             `json:"batch,omitempty"`
	Labels         Labels             `json:"labels,omitempty"`
	Source         TaskSource         `json:"source,omitempty"`
	Retries        int                `json:"retries"`
	Retried        int                `json:"retried"`
	Error          string             `json:"error,omitempty"`
	UUID           string             `json:"uuid"`
	CreatedAt      int64              `json:"createdAt"`
	Progress       float64            `json:"progress"`
	Priority       uint               `json:"priority"`
	StartedAt      int64              `json:"startedAt,omitempty"`
	FinishedAt     int64              `json:"finishedAt,omitempty"`
	Remaining      float64            `json:"remaining"`
	UpdatedAt      int64              `json:"updatedAt"`
}

type MetadataMap map[string]any
