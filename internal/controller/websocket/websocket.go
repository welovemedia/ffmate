package websocket

import (
	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/service"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/websocket"
)

type Service interface {
	Add(uuid string, c *websocket.Conn)
	Remove(uuid string)
}

type Controller struct {
	goyave.Component
	websocketService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	c.websocketService = server.Service(service.Websocket).(Service)
	debug.Controller.Debug("registered websocket controller")
}

func (c *Controller) OnUpgradeError(response *goyave.Response, _ *goyave.Request, status int, reason error) {
	message := map[string]string{
		"error": reason.Error(),
	}
	response.JSON(status, message)
}

func (ctrl *Controller) CheckOrigin(request *goyave.Request) bool {
	return true
}

func (w *Controller) Serve(c *websocket.Conn, request *goyave.Request) error {
	uuid := uuid.NewString()
	w.websocketService.Add(uuid, c)
	debug.Websocket.Debug("new connection from '%s' (uuid: %s)", request.RemoteAddress(), uuid)

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			debug.Websocket.Debug("disconnect from '%s' (uuid: %s): %v", request.RemoteAddress(), uuid, err)
			break
		}
	}

	return nil
}
