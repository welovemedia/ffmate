package dto

type TaskSource string

const (
	API         TaskSource = "api"
	WATCHFOLDER TaskSource = "watchfolder"
)

type TaskStatus string

const (
	QUEUED          TaskStatus = "QUEUED"
	RUNNING         TaskStatus = "RUNNING"
	PRE_PROCESSING  TaskStatus = "PRE_PROCESSING"
	POST_PROCESSING TaskStatus = "POST_PROCESSING"
	DONE_SUCCESSFUL TaskStatus = "DONE_SUCCESSFUL"
	DONE_ERROR      TaskStatus = "DONE_ERROR"
	DONE_CANCELED   TaskStatus = "DONE_CANCELED"
)

type NewTask struct {
	Command string `json:"command"`
	Preset  string `json:"preset"`

	Name string `json:"name"`

	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`

	Metadata *MetadataMap `json:"metadata"`

	Priority uint `json:"priority"`

	Webhooks *DirectWebhooks `json:"webhooks"`

	PreProcessing  *NewPrePostProcessing `json:"preProcessing"`
	PostProcessing *NewPrePostProcessing `json:"postProcessing"`
}

type Task struct {
	Uuid  string `json:"uuid"`
	Batch string `json:"batch,omitempty"`

	Name string `json:"name,omitempty"`

	Command    *RawResolved `json:"command"`
	InputFile  *RawResolved `json:"inputFile"`
	OutputFile *RawResolved `json:"outputFile"`

	Metadata *MetadataMap `json:"metadata,omitempty"` // Additional metadata for the task

	Status    TaskStatus `json:"status"`
	Progress  float64    `json:"progress"`
	Remaining float64    `json:"remaining"`

	Error string `json:"error,omitempty"`

	Priority uint `json:"priority"`

	Source TaskSource `json:"source,omitempty"`

	Webhooks *DirectWebhooks `json:"webhooks,omitempty"`

	PreProcessing  *PrePostProcessing `json:"preProcessing,omitempty"`
	PostProcessing *PrePostProcessing `json:"postProcessing,omitempty"`

	Client *Client `json:"client,omitempty"`

	StartedAt  int64 `json:"startedAt,omitempty"`
	FinishedAt int64 `json:"finishedAt,omitempty"`

	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
}

type MetadataMap map[string]any
