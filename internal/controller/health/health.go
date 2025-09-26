package health

import (
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"goyave.dev/goyave/v5"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered health controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Get("/health", c.get)
}

func (c *Controller) get(response *goyave.Response, _ *goyave.Request) {
	status := &dto.Health{Status: dto.HealthError}
	statusCode := 500
	if c.Server().IsReady() {
		status.Status = dto.HealthOk
		statusCode = 200
	}
	response.JSON(statusCode, status)
}
