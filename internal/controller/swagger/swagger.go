package swagger

import (
	"fmt"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/docs"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/auth"
)

type Controller struct {
	goyave.Component
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	debug.Controller.Debug("registered swagger controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	docs.SwaggerInfo.Version = c.Server().Config().GetString("app.version")
	docs.SwaggerInfo.Host += fmt.Sprintf(":%d", c.Server().Config().GetInt("server.port"))
	docs.SwaggerInfo.Schemes = []string{"http"}
	// public: reachable without authentication
	router.Get("/swagger/{*}", func(response *goyave.Response, request *goyave.Request) {
		httpSwagger.WrapHandler.ServeHTTP(response, request.Request())
	}).SetMeta(auth.MetaAuth, false)

	router.Get("/swagger", func(response *goyave.Response, request *goyave.Request) {
		http.Redirect(response, request.Request(), "/swagger/index.html", http.StatusPermanentRedirect)
	}).SetMeta(auth.MetaAuth, false)
	router.Get("/swagger/", func(response *goyave.Response, request *goyave.Request) {
		http.Redirect(response, request.Request(), "/swagger/index.html", http.StatusPermanentRedirect)
	}).SetMeta(auth.MetaAuth, false)
}
