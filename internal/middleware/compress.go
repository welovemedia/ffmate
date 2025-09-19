package middleware

import (
	"compress/gzip"
	"strings"

	"github.com/welovemedia/ffmate/internal/debug"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/middleware/compress"
)

type CompressMiddleware struct {
	goyave.Component
}

func (m *CompressMiddleware) Init(_ *goyave.Server) {
	debug.Middleware.Debug("registered compress middleware")
}

func (m *CompressMiddleware) Handle(next goyave.Handler) goyave.Handler {
	// exclude /swagger/* from compression as httpSwagger drops the Content-Encoding header
	return func(response *goyave.Response, request *goyave.Request) {
		path := request.Request().URL.Path

		if strings.HasPrefix(path, "/swagger") {
			next(response, request)
			return
		}

		comress := &compress.Middleware{
			Encoders: []compress.Encoder{
				&compress.Brotli{Quality: 11},
				&compress.Gzip{Level: gzip.BestCompression},
				&compress.LZW{},
			},
		}

		compressedHandler := comress.Handle(next)
		compressedHandler(response, request)
	}
}
