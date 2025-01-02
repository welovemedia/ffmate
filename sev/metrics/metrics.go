package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var gauges = make(map[string]prometheus.Gauge)
var dummyGauge = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dummy", Help: "Dummy gauge"})

type Metrics struct {
	Registry *prometheus.Registry
	Logger   *logrus.Logger
}

func (m *Metrics) Init() {
	m.Registry = prometheus.NewRegistry()
}

func (m *Metrics) RegisterGauge(name string, gauge prometheus.Gauge) {
	gauges[name] = gauge
	m.Registry.MustRegister(gauge)
	m.Logger.Debugf("registered prometheus gauge '%s'", name)
}

func (m *Metrics) Gauge(name string) prometheus.Gauge {
	if gauge, ok := gauges[name]; ok {
		return gauge
	}
	m.Logger.Warnf("prometheus gauge '%s' not found, returned dummy", name)
	return dummyGauge // return dummy gauge if gauge not found so .Inc() will never fail
}

func (m *Metrics) Gauges() map[string]prometheus.Gauge {
	return gauges
}
