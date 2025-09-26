package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite"
)

func TestClient(t *testing.T) {
	server := testsuite.InitServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
	response := server.TestRequest(request)
	body, _ := testsuite.ParseJsonBody[[]dto.Client](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/clients")
	assert.Equal(t, len(body), 1, "GET /api/v1/clients")
	assert.Equal(t, body[0].Session, cfg.GetString("ffmate.session"), "GET /api/v1/clients")
	assert.True(t, body[0].Self, "GET /api/v1/clients")
}
