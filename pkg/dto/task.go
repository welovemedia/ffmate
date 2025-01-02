package dto

import (
	"time"
)

type TaskStatus string

const (
	QUEUED          TaskStatus = "QUEUED"
	RUNNING         TaskStatus = "RUNNING"
	DONE_SUCCESSFUL TaskStatus = "DONE_SUCCESSFUL"
	DONE_ERROR      TaskStatus = "DONE_ERROR"
	DONE_CANCELED   TaskStatus = "DONE_CANCELED"
)

type Task struct {
	Uuid string `json:"uuid"`

	Command    string `json:"command"`
	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`

	Status   TaskStatus `json:"status"`
	Progress uint       `json:"progress"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
