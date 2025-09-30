package middleware

import (
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/auth"
)

func Register(_ *goyave.Server, router *goyave.Router) {
	if cfg.GetBool("ffmate.isAuth") {
		router.Middleware(ConfigCustomBasicAuth()).SetMeta(auth.MetaAuth, true)
	}

	router.Middleware(
		&CompressMiddleware{},
		&DebugoMiddleware{},
		&VersionMiddleware{},
	)
}
