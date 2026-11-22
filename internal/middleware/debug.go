package middleware

import (
	"strings"

	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
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
		if strings.HasPrefix(request.URL().Path, "/api/") {
			metrics.GaugeVec("rest.api").WithLabelValues(request.Method(), metricsPath(request)).Inc()
		}

		next(response, request)
	}
}

// metricsPath returns the route template (e.g. /api/v1/tasks/{uuid}) instead of the raw request path,
// so the number of metric series stays bounded. Requests not matching any route (normally not passed
// through router middleware at all) share one label.
func metricsPath(request *goyave.Request) string {
	route := request.Route
	if route == nil || route.GetName() == goyave.RouteNotFound || route.GetName() == goyave.RouteMethodNotAllowed {
		return "unmatched"
	}
	return route.GetFullURI()
}
