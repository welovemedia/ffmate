package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/welovemedia/ffmate/v2/internal/dialect"
	"github.com/welovemedia/ffmate/v2/testsuite"
)

func TestPrometheus(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodGet, "/metrics", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	body, _ := testsuite.ParseBody(response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /metrics")
	assert.Containsf(t, string(body), "ffmate_", "GET /metrics")
}
