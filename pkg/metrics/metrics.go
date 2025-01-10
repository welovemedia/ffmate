package metrics

import "github.com/prometheus/client_golang/prometheus"

var namespace = "ffmate"

var gauges = map[string]prometheus.Gauge{
	"task.created":        prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_created", Help: "Number of created tasks"}),
	"task.deleted":        prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_deleted", Help: "Number of deleted tasks"}),
	"task.status.updated": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_status_updated", Help: "Number of updated tasks"}),

	"preset.created": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "preset_created", Help: "Number of created presets"}),
	"preset.deleted": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "preset_deleted", Help: "Number of deleted presets"}),

	"webhook.created":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_created", Help: "Number of created webhooks"}),
	"webhook.executed": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_executed", Help: "Number of executed webhooks"}),
	"webhook.deleted":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_deleted", Help: "Number of deleted webhooks"}),
}

type MetricsImpl interface {
	Gauges() []prometheus.Gauge
}

type Metrics struct {
	MetricsImpl
}

func (m *Metrics) Gauges() map[string]prometheus.Gauge {
	return gauges
}
