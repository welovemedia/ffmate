package umami

import (
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/metrics"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/typeutil"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered umami controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Post("/umami", c.add)
}

func (c *Controller) add(response *goyave.Response, request *goyave.Request) {
	a := typeutil.MustConvert[dto.Umami](request.Data)
	metrics.GaugeVec("umami").WithLabelValues(a.Payload.Url, a.Payload.Screen, a.Payload.Langugage).Inc()
	response.Status(204)
}
