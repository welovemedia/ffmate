package middleware

import (
	"strings"

	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/metrics"
	"goyave.dev/goyave/v5"
)

type DebugoMiddleware struct {
	goyave.Component
}

func (m *DebugoMiddleware) Init(_ *goyave.Server) {
	debug.Middleware.Debug("registered debug middleware")
}

func (m *DebugoMiddleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		debug.HTTP.Debug("%s \"%s\"", request.Method(), request.URL())

		// add metrics for /api/* paths
		path := request.URL().Path
		if strings.HasPrefix(path, "/api/") {
			metrics.GaugeVec("rest.api").WithLabelValues(request.Method(), path).Inc()
		}

		next(response, request)
	}
}
