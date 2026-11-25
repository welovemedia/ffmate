package health

import (
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/auth"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered health controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	// public: reachable without authentication
	router.Get("/health", c.get).SetMeta(auth.MetaAuth, false)
}

func (c *Controller) get(response *goyave.Response, _ *goyave.Request) {
	status := &dto.Health{Status: dto.HealthOk}
	statusCode := 200
	if !c.Server().IsReady() || !c.Server().HasDB() {
		status.Status = dto.HealthError
		statusCode = 500
	}
	response.JSON(statusCode, status)
}
