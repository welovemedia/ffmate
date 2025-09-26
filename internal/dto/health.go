package dto

type HealthStatus string

const (
	HealthOk    HealthStatus = "ok"
	HealthError HealthStatus = "error"
)

type Health struct {
	Status HealthStatus `json:"status"`
}
