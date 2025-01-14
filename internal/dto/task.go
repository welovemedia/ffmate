package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type TaskStatus string

const (
	QUEUED          TaskStatus = "QUEUED"
	RUNNING         TaskStatus = "RUNNING"
	POST_PROCESSING TaskStatus = "POST_PROCESSING"
	DONE_SUCCESSFUL TaskStatus = "DONE_SUCCESSFUL"
	DONE_ERROR      TaskStatus = "DONE_ERROR"
	DONE_CANCELED   TaskStatus = "DONE_CANCELED"
)

type Resolved struct {
	Command        string `json:"command"`
	InputFile      string `json:"inputFile"`
	OutputFile     string `json:"outputFile"`
	PostProcessing struct {
		ScriptPath  string `json:"scriptPath"`
		SidecarPath string `json:"sidecarPath"`
	} `json:"postProcessing"`
}

type PostProcessing struct {
	ScriptPath  string `json:"scriptPath"`
	SidecarPath string `json:"sidecarPath"`
	Error       string `json:"error,omitempty"`
	StartedAt   int64  `json:"startedAt,omitempty"`
	FinishedAt  int64  `json:"finishedAt,omitempty"`
}

type Task struct {
	Uuid  string `json:"uuid"`
	Batch string `json:"batch,omitempty"`

	Name string `json:"name,omitempty"`

	Resolved *Resolved `json:"resolved,omitempty"`

	Command    string `json:"command"`
	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`

	Status   TaskStatus `json:"status"`
	Progress float64    `json:"progress"`
	Error    string     `json:"error,omitempty"`

	Priority uint `json:"priority"`

	PostProcessing *PostProcessing `json:"postProcessing,omitempty"`

	StartedAt  int64 `json:"startedAt,omitempty"`
	FinishedAt int64 `json:"finishedAt,omitempty"`

	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
}

func (p PostProcessing) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *PostProcessing) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

func (p Resolved) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Resolved) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}
