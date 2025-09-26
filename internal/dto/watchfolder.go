package dto

type NewWatchfolder struct {
	Filter       *WatchfolderFilter `json:"filter"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Path         string             `json:"path"`
	Preset       string             `json:"preset"`
	Interval     int                `json:"interval"`
	GrowthChecks int                `json:"growthChecks"`
	Suspended    bool               `json:"suspended"`
}

type Watchfolder struct {
	Filter       *WatchfolderFilter `json:"filter"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Path         string             `json:"path"`
	Error        string             `json:"error,omitempty"`
	UUID         string             `json:"uuid"`
	Preset       string             `json:"preset"`
	CreatedAt    int64              `json:"createdAt"`
	GrowthChecks int                `json:"growthChecks"`
	UpdatedAt    int64              `json:"updatedAt"`
	Interval     int                `json:"interval"`
	LastCheck    int64              `json:"lastCheck"`
	Suspended    bool               `json:"suspended"`
}

type WatchfolderFilter struct {
	Extensions *WatchfolderFilterExtensions `json:"extensions"`
}

type WatchfolderFilterExtensions struct {
	Exclude []string `json:"exclude"`
	Include []string `json:"include"`
}
