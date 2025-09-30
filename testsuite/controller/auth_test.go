package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestAuthSuccess(t *testing.T) {
	cfg.Set("ffmate.isAuth", true)
	server := testsuite.InitServer(t)
	request := testsuite.NewRequest(http.MethodGet, "/api/v1/version", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/version")
	cfg.Set("ffmate.isAuth", false)
}

func TestAuthWrongCredentials(t *testing.T) {
	cfg.Set("ffmate.isAuth", true)
	server := testsuite.InitServer(t)
	request := testsuite.NewRequest(http.MethodGet, "/api/v1/version", nil)
	request.Header.Set("Authorization", "Basic wrong.credentials")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "POST /api/v1/version")
	cfg.Set("ffmate.isAuth", false)
}

func TestAuthNoCredentials(t *testing.T) {
	cfg.Set("ffmate.isAuth", true)
	server := testsuite.InitServer(t)
	request := testsuite.NewRequest(http.MethodGet, "/api/v1/version", nil)
	request.Header.Del("Authorization")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "POST /api/v1/version")
	cfg.Set("ffmate.isAuth", false)
}
