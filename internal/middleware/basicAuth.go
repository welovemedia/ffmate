package middleware

import (
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/auth"
)

type CustomConfigBasicAuthenticator struct {
	auth.Authenticator[auth.BasicUser] // interface
}

func (a *CustomConfigBasicAuthenticator) OnUnauthorized(response *goyave.Response, _ *goyave.Request, err error) {
	response.Header().Add("WWW-Authenticate", `Basic realm="ffmate"`)
	response.JSON(401, map[string]any{
		"error": err.Error(),
	})
}

func ConfigCustomBasicAuth() *auth.Handler[auth.BasicUser] {
	handler := auth.ConfigBasicAuth()

	handler.Authenticator = &CustomConfigBasicAuthenticator{
		Authenticator: auth.ConfigBasicAuth(),
	}

	return handler
}
