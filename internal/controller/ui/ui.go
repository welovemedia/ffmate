package ui

import (
	"embed"
	_ "embed"
	"io/fs"
	"net/http"

	"github.com/welovemedia/ffmate/v2/internal/debug"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/fsutil"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered ui controller")
}

//go:embed all:ui-build/*
var uiEmbed embed.FS

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	// redirect / to /ui
	router.Get("/", func(response *goyave.Response, request *goyave.Request) {
		http.Redirect(response, request.Request(), "/ui", http.StatusPermanentRedirect)
	})

	sub, err := fs.Sub(uiEmbed, "ui-build")
	if err != nil {
		panic(err)
	}
	router.Static(fsutil.NewEmbed(sub.(fs.ReadDirFS)), "/ui", false)
}
