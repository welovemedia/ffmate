package dto

type Umami struct {
	Type    string       `json:"type"`
	Payload UmamiPayload `json:"payload"`
}

type UmamiPayload struct {
	Hostname  string `json:"hostname"`
	Langugage string `json:"language"`
	Screen    string `json:"screen"`
	URL       string `json:"url"`
	Referrer  string `json:"referrer"`
	Title     string `json:"title"`
	Website   string `json:"website"`
}
