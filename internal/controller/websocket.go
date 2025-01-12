package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/sev"
	"github.com/yosev/debugo"
)

var debug = debugo.New("websocket:controller")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(g *http.Request) bool {
		return true
	},
}

type WebsocketController struct {
	sev.Controller
	sev              *sev.Sev
	websocketService *service.WebsocketService

	Prefix string
}

func (c *WebsocketController) Setup(s *sev.Sev) {
	c.websocketService = &service.WebsocketService{}
	c.sev = s
	s.Gin().GET(c.Prefix+c.getEndpoint(), c.websocket)
}

func (c *WebsocketController) websocket(gin *gin.Context) {
	conn, err := upgrader.Upgrade(gin.Writer, gin.Request, nil)
	if err != nil {
		c.sev.Logger().Errorf("failed to establish websocket connection: %v", err)
		return
	}
	uuid := uuid.NewString()

	defer conn.Close()
	defer c.websocketService.RemoveConnection(uuid, conn)

	c.websocketService.AddConnection(uuid, conn)

	debug.Debugf("new connection from %s (uuid: %s)", gin.RemoteIP(), uuid)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			debug.Debugf("disconnection from %s: %v", gin.RemoteIP(), err)
			break
		}
	}
}

func (c *WebsocketController) GetName() string {
	return "webhook"
}

func (c *WebsocketController) getEndpoint() string {
	return "/v1/ws"
}
