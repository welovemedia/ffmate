package debug

import (
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/yosev/debugo"
	"goyave.dev/goyave/v5"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered debug controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Delete("/debug", c.delete)
	router.Patch("/debug/{namespace}", c.set)
}

// @Summary Set debug namespace(s)
// @Description Set debug namespace(s)
// @Tags debug
// @Success 204
// @Router /debug/{namespace} [patch]
func (c *Controller) set(response *goyave.Response, request *goyave.Request) {
	ns := request.RouteParams["namespace"]
	debugo.SetNamespace(ns)
	debug.Log.Info("changed debug namespace to '%s'", ns)

	response.Status(204)
}

// @Summary Turn debugging off
// @Description Turn debugging off
// @Tags debug
// @Success 204
// @Router /debug [delete]
func (c *Controller) delete(response *goyave.Response, request *goyave.Request) {
	debug.Log.Info("disabled debug namespace")
	debugo.SetNamespace("")
	response.Status(204)
}
