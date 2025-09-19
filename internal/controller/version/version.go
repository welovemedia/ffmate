package version

import (
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/dto"
	"goyave.dev/goyave/v5"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered version controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Get("/version", c.get)
}

func (c *Controller) get(response *goyave.Response, request *goyave.Request) {
	response.JSON(200, &dto.Version{Version: c.Config().GetString("app.version")})
}
