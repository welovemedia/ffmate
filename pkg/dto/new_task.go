package dto

type NewTask struct {
	Command    string `json:"command"`
	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`
}
