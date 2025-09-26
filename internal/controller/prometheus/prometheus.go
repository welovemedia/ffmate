package prometheus

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"goyave.dev/goyave/v5"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered prometheus controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Get("/metrics", func(response *goyave.Response, request *goyave.Request) {
		handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})
		handler.ServeHTTP(response, request.Request())
	})
}
