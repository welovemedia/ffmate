package middleware

import (
	"errors"
	"net/http"

	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/exception"
	"goyave.dev/goyave/v5"
)

var errBasicAuthDisabled = errors.New("this route requires basic auth to be enabled")

// RequireBasicAuthMiddleware rejects all requests with 412 if basic auth is disabled.
// Register it on a (sub)router so it runs before route level validation:
//
//	router.Subrouter("/apikeys").Middleware(&middleware.RequireBasicAuthMiddleware{})
type RequireBasicAuthMiddleware struct {
	goyave.Component
}

func (m *RequireBasicAuthMiddleware) Init(server *goyave.Server) {
	m.Component.Init(server)
	debug.Middleware.Debug("registered require basic auth middleware")
}

func (m *RequireBasicAuthMiddleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		if !cfg.GetBool("ffmate.isAuth") {
			response.JSON(http.StatusPreconditionFailed, exception.HTTPPreconditionFailed(errBasicAuthDisabled, ""))
			return
		}
		next(response, request)
	}
}
