package sev

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/welovemedia/ffmate/sev/metrics"
)

var debugMetrics = debug.Extend("metrics")

func (s *Sev) Metrics() *metrics.Metrics {
	return s.metrics
}

func (s *Sev) registerMetrics() {
	h := promhttp.HandlerFor(s.metrics.Registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
	s.Gin().GET("/metrics", func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	})
	debugMetrics.Debug("registered prometheus http handler")
}
