package controller

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "github.com/welovemedia/ffmate/v2/internal/dialect"
)

func gaugeValue(t *testing.T, g prometheus.Gauge) float64 {
	m := &dto.Metric{}
	assert.NoError(t, g.Write(m))
	return m.GetGauge().GetValue()
}

func seriesCount(c prometheus.Collector) int {
	ch := make(chan prometheus.Metric, 1000)
	c.Collect(ch)
	close(ch)
	return len(ch)
}

func TestRestAPIMetricUsesRouteTemplate(t *testing.T) {
	server := testsuite.InitServer(t)
	vec := metrics.GaugeVec("rest.api")
	route := vec.WithLabelValues(http.MethodGet, "/api/v1/tasks/{uuid}")
	routeBefore := gaugeValue(t, route)
	seriesBefore := seriesCount(vec)

	get := func(path string) {
		response := server.TestRequest(testsuite.NewRequest(http.MethodGet, path, nil))
		_ = response.Body.Close()
	}
	for i := 0; i < 3; i++ {
		get("/api/v1/tasks/" + uuid.NewString())
	}
	get("/api/v1/tasks/not-a-uuid")
	get("/api/v1/does-not-exist/" + uuid.NewString())
	get("/api/v1/does-not-exist/" + uuid.NewString())

	assert.InDelta(t, routeBefore+4, gaugeValue(t, route), 0.001, "all task lookups (also invalid uuids) share the route template label")
	assert.Equal(t, seriesBefore, seriesCount(vec), "no new series per uuid or unknown path")
}
