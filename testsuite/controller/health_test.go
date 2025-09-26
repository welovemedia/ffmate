package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite"

	"goyave.dev/goyave/v5"
	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestHealth(t *testing.T) {
	server := testsuite.InitServer(t)

	// expect error
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	body, _ := testsuite.ParseJSONBody[dto.Health](response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode, "GET /health")
	assert.Equal(t, dto.HealthError, body.Status, "GET /health")

	server.RegisterStartupHook(func(*goyave.Server) {
		// expect ok
		request = httptest.NewRequest(http.MethodGet, "/health", nil)
		resp := server.TestRequest(request)
		defer resp.Body.Close() // nolint:errcheck
		body, _ = testsuite.ParseJSONBody[dto.Health](resp.Body)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "GET /health")
		assert.Equal(t, dto.HealthOk, body.Status, "GET /health")
		server.Stop()
	})

	_ = server.Start()
}
