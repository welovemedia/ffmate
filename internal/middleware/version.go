package middleware

import (
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"goyave.dev/goyave/v5"
)

type VersionMiddleware struct {
	goyave.Component
}

func (m *VersionMiddleware) Init(server *goyave.Server) {
	m.Component.Init(server)
	debug.Middleware.Debug("registered version middleware")
}

func (m *VersionMiddleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		response.Header().Set("X-Server", m.Config().GetString("app.name")+"/v"+m.Config().GetString("app.version"))
		next(response, request)
	}
}
