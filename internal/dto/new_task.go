package dto

type NewTask struct {
	Command string `json:"command"`
	Preset  string `json:"preset"`

	Name string `json:"name"`

	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`

	Metadata *InterfaceMap `json:"metadata,omitempty"` // Additional metadata for the task

	Priority uint `json:"priority"`

	PreProcessing  *NewPrePostProcessing `json:"preProcessing"`
	PostProcessing *NewPrePostProcessing `json:"postProcessing"`
}
