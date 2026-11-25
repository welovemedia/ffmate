package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "github.com/welovemedia/ffmate/v2/internal/dialect"
)

func TestAuthSuccess(t *testing.T) {
	cfg.Set("ffmate.isAuth", true)
	server := testsuite.InitServer(t)
	request := testsuite.NewRequest(http.MethodGet, "/api/v1/presets", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/presets")
	cfg.Set("ffmate.isAuth", false)
}

func TestAuthWrongCredentials(t *testing.T) {
	cfg.Set("ffmate.isAuth", true)
	server := testsuite.InitServer(t)
	request := testsuite.NewRequest(http.MethodGet, "/api/v1/presets", nil)
	request.Header.Set("Authorization", "Basic wrong.credentials")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "GET /api/v1/presets")
	cfg.Set("ffmate.isAuth", false)
}

func TestAuthNoCredentials(t *testing.T) {
	cfg.Set("ffmate.isAuth", true)
	server := testsuite.InitServer(t)
	request := testsuite.NewRequest(http.MethodGet, "/api/v1/presets", nil)
	request.Header.Del("Authorization")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "GET /api/v1/presets")
	cfg.Set("ffmate.isAuth", false)
}

func TestAuthPublicRoutes(t *testing.T) {
	cfg.Set("ffmate.isAuth", true)
	defer cfg.Set("ffmate.isAuth", false)
	server := testsuite.InitServer(t)

	for _, path := range []string{"/api/v1/version", "/health", "/swagger/index.html"} {
		request := testsuite.NewRequest(http.MethodGet, path, nil)
		request.Header.Del("Authorization")
		response := server.TestRequest(request)
		response.Body.Close() // nolint:errcheck
		assert.NotEqual(t, http.StatusUnauthorized, response.StatusCode, "GET "+path)
	}
}
