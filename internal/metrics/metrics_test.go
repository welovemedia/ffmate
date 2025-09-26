package metrics

import (
	"testing"
)

func TestGauge(t *testing.T) {
	for name := range Gauges() {
		Gauge(name).Inc()
	}
}

func TestGaugeVec(t *testing.T) {
	for name, g := range GaugesVec() {
		labels, ok := gaugeVecLabels[name]
		if !ok {
			t.Fatalf("no label names defined for gaugeVec %q", name)
		}

		values := make([]string, len(labels))
		for i, ln := range labels {
			values[i] = "test_" + ln
		}

		g.WithLabelValues(values...).Inc()
	}
}