package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestVersionGet(t *testing.T) {
	server := testsuite.InitServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	response := server.TestRequest(request)
	version, _ := testsuite.ParseJsonBody[dto.Version](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/webhooks")
	assert.Equal(t, version.Version, "test-1.0.0", "POST /api/v1/webhooks")
	assert.Equal(t, response.Header.Get("X-Server"), "ffmate/vtest-1.0.0", "POST /api/v1/webhooks")
}
