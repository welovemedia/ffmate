package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/testsuite"
	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestSwagger(t *testing.T) {
	server := testsuite.InitServer(t)

	request := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	response := server.TestRequest(request)
	assert.Equal(t, response.Header.Get("Location"), "/swagger/index.html", "GET /swagger")
	assert.Equal(t, http.StatusPermanentRedirect, response.StatusCode, "GET /swagger")

	request = httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	response = server.TestRequest(request)
	assert.Equal(t, response.Header.Get("Location"), "/swagger/index.html", "GET /swagger/")
	assert.Equal(t, http.StatusPermanentRedirect, response.StatusCode, "GET /swagger/")

	request = httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	response = server.TestRequest(request)
	body, _ := testsuite.ParseBody(response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /swagger/index.html")
	assert.Containsf(t, string(body), "swagger-ui.css", "GET /swagger/index.html")
}
