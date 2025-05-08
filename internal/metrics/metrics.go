package metrics

import "github.com/prometheus/client_golang/prometheus"

var namespace = "ffmate"

var gauges = map[string]prometheus.Gauge{
	"batch.created":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "batch_created", Help: "Number of created batches"}),
	"batch.finished": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "batch_finished", Help: "Number of finished batches"}),

	"task.created":   prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_created", Help: "Number of created tasks"}),
	"task.deleted":   prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_deleted", Help: "Number of deleted tasks"}),
	"task.updated":   prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_updated", Help: "Number of updated tasks"}),
	"task.canceled":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_canceled", Help: "Number of canceled tasks"}),
	"task.restarted": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "task_restarted", Help: "Number of restarted tasks"}),

	"preset.created": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "preset_created", Help: "Number of created presets"}),
	"preset.deleted": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "preset_deleted", Help: "Number of deleted presets"}),

	"webhook.created":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_created", Help: "Number of created webhooks"}),
	"webhook.executed": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_executed", Help: "Number of executed webhooks"}),
	"webhook.deleted":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_deleted", Help: "Number of deleted webhooks"}),

	"watchfolder.created":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_created", Help: "Number of created watchfolders"}),
	"watchfolder.executed": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_executed", Help: "Number of executed watchfolders"}),
	"watchfolder.updated":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_updated", Help: "Number of updated watchfolder"}),
	"watchfolder.deleted":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_deleted", Help: "Number of deleted watchfolders"}),
}

var gaugesVec = map[string]*prometheus.GaugeVec{
	"rest.api": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "rest_api",
			Help:      "Number of requests against the RestAPI",
		},
		[]string{"method", "path"},
	),
	"umami": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "umami",
			Help:      "Number of requests coming from umami",
		},
		[]string{"url", "screen", "language"},
	),
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

func (m *Metrics) GaugesVec() map[string]*prometheus.GaugeVec {
	return gaugesVec
}
