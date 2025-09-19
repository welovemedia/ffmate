package middleware

import (
	"goyave.dev/goyave/v5"
)

func Register(_ *goyave.Server, router *goyave.Router) {
	router.Middleware(
		&CompressMiddleware{},
		&DebugoMiddleware{},
		&VersionMiddleware{},
	)
}
