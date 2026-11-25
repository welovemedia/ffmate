package middleware

import (
	"errors"
	"net/http"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/auth"
)

const HeaderAPIKey = "X-API-Key"

var (
	errInvalidAPIKey      = errors.New("invalid api key")
	errInsufficientAccess = errors.New("api key has insufficient access for this route")
)

type APIKeyResolver interface {
	Resolve(key string) (*model.APIKey, error)
}

// CustomConfigBasicAuthenticator authenticates a request either by an api key
// (X-API-Key header) or, if no api key is given, by the configured basic auth credentials
type CustomConfigBasicAuthenticator struct {
	auth.Authenticator[auth.BasicUser]
	apiKeys APIKeyResolver
}

func (a *CustomConfigBasicAuthenticator) Init(server *goyave.Server) {
	a.Authenticator.Init(server)
	a.apiKeys = server.Service(service.APIKey).(APIKeyResolver)
}

func (a *CustomConfigBasicAuthenticator) Authenticate(request *goyave.Request) (*auth.BasicUser, error) {
	key := request.Header().Get(HeaderAPIKey)
	if key == "" {
		return a.Authenticator.Authenticate(request)
	}

	apiKey, err := a.apiKeys.Resolve(key)
	if err != nil {
		panic(err) // unexpected (database) error, results in a 500
	}
	if apiKey == nil {
		return nil, errInvalidAPIKey
	}
	if !hasAccess(request.Route, apiKey.Access) {
		return nil, errInsufficientAccess
	}

	return &auth.BasicUser{Name: "apikey:" + apiKey.UUID}, nil
}

func (a *CustomConfigBasicAuthenticator) OnUnauthorized(response *goyave.Response, _ *goyave.Request, err error) {
	if errors.Is(err, errInsufficientAccess) {
		response.JSON(http.StatusForbidden, map[string]any{
			"error": err.Error(),
		})
		return
	}

	response.Header().Add("WWW-Authenticate", `Basic realm="ffmate"`)
	response.JSON(http.StatusUnauthorized, map[string]any{
		"error": err.Error(),
	})
}

func ConfigCustomBasicAuth() *auth.Handler[auth.BasicUser] {
	return auth.Middleware[auth.BasicUser](&CustomConfigBasicAuthenticator{
		Authenticator: &auth.ConfigBasicAuthenticator{},
	})
}
