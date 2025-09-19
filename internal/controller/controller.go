package controller

import (
	"github.com/welovemedia/ffmate/internal/controller/client"
	"github.com/welovemedia/ffmate/internal/controller/debug"
	"github.com/welovemedia/ffmate/internal/controller/health"
	"github.com/welovemedia/ffmate/internal/controller/preset"
	"github.com/welovemedia/ffmate/internal/controller/prometheus"
	"github.com/welovemedia/ffmate/internal/controller/settings"
	"github.com/welovemedia/ffmate/internal/controller/swagger"
	"github.com/welovemedia/ffmate/internal/controller/task"
	"github.com/welovemedia/ffmate/internal/controller/ui"
	"github.com/welovemedia/ffmate/internal/controller/umami"
	"github.com/welovemedia/ffmate/internal/controller/version"
	"github.com/welovemedia/ffmate/internal/controller/watchfolder"
	"github.com/welovemedia/ffmate/internal/controller/webhook"
	"github.com/welovemedia/ffmate/internal/controller/websocket"
	"goyave.dev/goyave/v5"
	ws "goyave.dev/goyave/v5/websocket"
)

func Register(server *goyave.Server, router *goyave.Router) {
	// service the UI
	router.Controller(&ui.Controller{})

	apiRouter := router.Subrouter("/api/v1")

	apiRouter.Controller(&version.Controller{})
	apiRouter.Controller(&preset.Controller{})
	apiRouter.Controller(&webhook.Controller{})
	apiRouter.Controller(&watchfolder.Controller{})
	apiRouter.Controller(&task.Controller{})
	apiRouter.Controller(&settings.Controller{})
	apiRouter.Controller(&client.Controller{})
	apiRouter.Controller(&debug.Controller{})

	// health
	router.Controller(&health.Controller{})

	// websocket
	apiRouter.Subrouter("/ws").Controller(ws.New(&websocket.Controller{}))

	// ui
	router.Controller(&ui.Controller{})

	// umami
	router.Controller(&umami.Controller{})

	// swagger
	router.Controller(&swagger.Controller{})

	// prometheus
	router.Controller(&prometheus.Controller{})
}
