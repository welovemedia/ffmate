package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite"
)

func TestClient(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/clients", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	body, _ := testsuite.ParseJSONBody[[]dto.Client](response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/clients")
	assert.Len(t, body, 1, "GET /api/v1/clients")
	assert.Equal(t, body[0].Session, cfg.GetString("ffmate.session"), "GET /api/v1/clients")
	assert.Contains(t, body[0].Labels, "test-label-1", "GET /api/v1/clients")
	assert.Contains(t, body[0].Labels, "test-label-2", "GET /api/v1/clients")
	assert.Contains(t, body[0].Labels, "test-label-3", "GET /api/v1/clients")
	assert.NotContains(t, body[0].Labels, "test-label-0", "GET /api/v1/clients")
	assert.Equal(t, body[0].MaxConcurrentTasks, cfg.GetInt("ffmate.maxConcurrentTasks"), "GET /api/v1/clients")
	assert.True(t, body[0].Self, "GET /api/v1/clients")
}
