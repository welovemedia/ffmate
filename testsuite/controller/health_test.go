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
	body, _ := testsuite.ParseJsonBody[dto.Health](response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode, "GET /health")
	assert.Equal(t, dto.HEALTH_ERROR, body.Status, "GET /health")

	server.RegisterStartupHook(func(*goyave.Server) {
		// expect ok
		request = httptest.NewRequest(http.MethodGet, "/health", nil)
		response = server.TestRequest(request)
		body, _ = testsuite.ParseJsonBody[dto.Health](response.Body)
		assert.Equal(t, http.StatusOK, response.StatusCode, "GET /health")
		assert.Equal(t, dto.HEALTH_OK, body.Status, "GET /health")
		server.Stop()
	})

	server.Start()

}
