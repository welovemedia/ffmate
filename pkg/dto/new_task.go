package dto

type NewTask struct {
	Command    string `json:"command"`
	Preset     string `json:"preset"`
	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`
}
