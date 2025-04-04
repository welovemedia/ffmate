package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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

type NewPrePostProcessing struct {
	ScriptPath  string `json:"scriptPath,omitempty"`
	SidecarPath string `json:"sidecarPath,omitempty"`
}

type PrePostProcessing struct {
	ScriptPath  *RawResolved `json:"scriptPath,omitempty"`
	SidecarPath *RawResolved `json:"sidecarPath,omitempty"`
	Error       string       `json:"error,omitempty"`
	StartedAt   int64        `json:"startedAt,omitempty"`
	FinishedAt  int64        `json:"finishedAt,omitempty"`
}

type RawResolved struct {
	Raw      string `json:"raw"`
	Resolved string `json:"resolved,omitempty"`
}

type Task struct {
	Uuid  string `json:"uuid"`
	Batch string `json:"batch,omitempty"`

	Name string `json:"name,omitempty"`

	Command    *RawResolved `json:"command"`
	InputFile  *RawResolved `json:"inputFile"`
	OutputFile *RawResolved `json:"outputFile"`

	Metadata *InterfaceMap `json:"metadata,omitempty"` // Additional metadata for the task

	Status    TaskStatus `json:"status"`
	Progress  float64    `json:"progress"`
	Remaining float64    `json:"remaining"`

	Error string `json:"error,omitempty"`

	Priority uint `json:"priority"`

	Source string `json:"source,omitempty"`

	PreProcessing  *PrePostProcessing `json:"preProcessing,omitempty"`
	PostProcessing *PrePostProcessing `json:"postProcessing,omitempty"`

	StartedAt  int64 `json:"startedAt,omitempty"`
	FinishedAt int64 `json:"finishedAt,omitempty"`

	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
}

type InterfaceMap map[string]interface{}

func (j InterfaceMap) Value() (interface{}, error) {
	return json.Marshal(j)
}

func (j *InterfaceMap) Scan(value interface{}) error {
	if value == nil {
		*j = InterfaceMap{}
		return nil
	}

	// Handle different types (DB drivers may return different types)
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("unsupported data type: %T", value)
	}
}

func (p PrePostProcessing) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *PrePostProcessing) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

func (p RawResolved) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *RawResolved) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

func (n NewPrePostProcessing) Value() (driver.Value, error) {
	return json.Marshal(n)
}

func (n *NewPrePostProcessing) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, n)
}
