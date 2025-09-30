package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/testsuite"
	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestUI(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodGet, "/", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, "/ui", response.Header.Get("Location"), "GET /ui")
	assert.Equal(t, http.StatusPermanentRedirect, response.StatusCode, "GET /ui")

	request = testsuite.NewRequest(http.MethodGet, "/ui", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /ui")

	request = testsuite.NewRequest(http.MethodGet, "/ui/index.html", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	body, _ := testsuite.ParseBody(response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /ui/index.html")
	assert.Containsf(t, string(body), "make build+frontend", "GET /ui/index.html")
}
