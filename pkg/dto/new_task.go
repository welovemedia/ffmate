package dto

type NewTask struct {
	Command string `json:"command"`
	Preset  string `json:"preset"`

	Name string `json:"name"`

	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`

	Priority uint `json:"priority"`
}
