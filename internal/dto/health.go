package dto

type HealthStatus string

const (
	HEALTH_OK    HealthStatus = "ok"
	HEALTH_ERROR HealthStatus = "error"
)

type Health struct {
	Status HealthStatus `json:"status"`
}
