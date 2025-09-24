package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type NewWatchfolder struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Path         string `json:"path"`
	Interval     int    `json:"interval"`
	GrowthChecks int    `json:"growthChecks"`

	Filter *WatchfolderFilter `json:"filter"`

	Labels Labels `json:"labels"`

	Suspended bool `json:"suspended"`

	Preset string `json:"preset"`
}

type Watchfolder struct {
	Uuid string `json:"uuid"`

	Name        string `json:"name"`
	Description string `json:"description"`

	Path         string `json:"path"`
	Interval     int    `json:"interval"`
	GrowthChecks int    `json:"growthChecks"`

	Suspended bool `json:"suspended"`

	Filter *WatchfolderFilter `json:"filter"`

	Preset string `json:"preset"`

	Labels Labels `json:"labels,omitempty"`

	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`

	Error     string `json:"error,omitempty"`
	LastCheck int64  `json:"lastCheck"`
}

type WatchfolderFilter struct {
	Extensions *WatchfolderFilterExtensions `json:"extensions"`
}

type WatchfolderFilterExtensions struct {
	Exclude []string `json:"exclude"`
	Include []string `json:"include"`
}

func (n WatchfolderFilter) Value() (driver.Value, error) {
	return json.Marshal(n)
}

func (n *WatchfolderFilter) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, n)
}
