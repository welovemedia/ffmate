package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/testsuite"
	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestUI(t *testing.T) {
	server := testsuite.InitServer(t)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := server.TestRequest(request)
	assert.Equal(t, response.Header.Get("Location"), "/ui", "GET /ui")
	assert.Equal(t, http.StatusPermanentRedirect, response.StatusCode, "GET /ui")

	request = httptest.NewRequest(http.MethodGet, "/ui", nil)
	response = server.TestRequest(request)
	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /ui")

	request = httptest.NewRequest(http.MethodGet, "/ui/index.html", nil)
	response = server.TestRequest(request)
	body, _ := testsuite.ParseBody(response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /ui/index.html")
	assert.Containsf(t, string(body), "make build+frontend", "GET /ui/index.html")
}
