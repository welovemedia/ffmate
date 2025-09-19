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
	"preset.updated": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "preset_updated", Help: "Number of updated presets"}),
	"preset.deleted": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "preset_deleted", Help: "Number of deleted presets"}),

	"webhook.created":         prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_created", Help: "Number of created webhooks"}),
	"webhook.executed":        prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_executed", Help: "Number of executed webhooks"}),
	"webhook.executed.direct": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_executed_direct", Help: "Number of directly executed webhooks"}),
	"webhook.updated":         prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_updated", Help: "Number of updated webhooks"}),
	"webhook.deleted":         prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "webhook_deleted", Help: "Number of deleted webhooks"}),

	"watchfolder.created":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_created", Help: "Number of created watchfolders"}),
	"watchfolder.executed": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_executed", Help: "Number of executed watchfolders"}),
	"watchfolder.updated":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_updated", Help: "Number of updated watchfolder"}),
	"watchfolder.deleted":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "watchfolder_deleted", Help: "Number of deleted watchfolders"}),

	"websocket.broadcast":  prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "websocket_broadcast", Help: "Number of broadcasted messages"}),
	"websocket.connect":    prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "websocket_connect", Help: "Number of websocket connections"}),
	"websocket.disconnect": prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: "websocket_disconnect", Help: "Number of websocket disconnections"}),
}

// use label map to make testing easier
var gaugeVecLabels = map[string][]string{
	"rest.api":            {"method", "path"},
	"umami":               {"url", "screen", "language"},
	"task.preProcessing":  {"sidecarPath", "scriptPath"},
	"task.postProcessing": {"sidecarPath", "scriptPath"},
	"preset.global":       {"name"},
}
var gaugesVec = map[string]*prometheus.GaugeVec{
	"rest.api": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "rest_api",
			Help:      "Number of requests against the RestAPI",
		},
		gaugeVecLabels["rest.api"],
	),
	"umami": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "umami",
			Help:      "Number of requests coming from umami",
		},
		gaugeVecLabels["umami"],
	),
	"task.preProcessing": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "task_preProcessing",
			Help:      "Number of preProcessing",
		},
		gaugeVecLabels["task.preProcessing"],
	),
	"task.postProcessing": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "task_postProcessing",
			Help:      "Number of postProcessing",
		},
		gaugeVecLabels["task.postProcessing"],
	),
	"preset.global": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "preset_global",
			Help:      "Number of global presets",
		},
		gaugeVecLabels["preset.global"],
	),
}

var Registry = prometheus.NewRegistry()

func init() {
	for _, gauge := range gauges {
		Registry.MustRegister(gauge)
	}
	for _, gaugeVec := range gaugesVec {
		Registry.MustRegister(gaugeVec)
	}
}

func Gauges() map[string]prometheus.Gauge {
	return gauges
}

func GaugesVec() map[string]*prometheus.GaugeVec {
	return gaugesVec
}

func Gauge(key string) prometheus.Gauge {
	if g, ok := gauges[key]; ok {
		return g
	}
	panic("gauge not found: " + key)
}

func GaugeVec(key string) *prometheus.GaugeVec {
	if g, ok := gaugesVec[key]; ok {
		return g
	}
	panic("gauge not found: " + key)
}
