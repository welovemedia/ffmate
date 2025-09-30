package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestVersionGet(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/version", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	version, _ := testsuite.ParseJSONBody[dto.Version](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/version")
	assert.Equal(t, "test-1.0.0", version.Version, "POST /api/v1/version")
	assert.Equal(t, "ffmate/vtest-1.0.0", response.Header.Get("X-Server"), "POST /api/v1/version")
}
