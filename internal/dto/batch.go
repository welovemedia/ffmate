package dto

type NewBatch struct {
	Tasks []*NewTask `json:"tasks"`
}

type Batch struct {
	Uuid  string  `json:"uuid"`
	Tasks []*Task `json:"tasks"`
}
