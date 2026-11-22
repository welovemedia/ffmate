package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/welovemedia/ffmate/v2/internal/dialect"
	"github.com/welovemedia/ffmate/v2/testsuite"
)

func TestSwagger(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodGet, "/swagger", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, "/swagger/index.html", response.Header.Get("Location"), "GET /swagger")
	assert.Equal(t, http.StatusPermanentRedirect, response.StatusCode, "GET /swagger")

	request = testsuite.NewRequest(http.MethodGet, "/swagger/", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, "/swagger/index.html", response.Header.Get("Location"), "GET /swagger/")
	assert.Equal(t, http.StatusPermanentRedirect, response.StatusCode, "GET /swagger/")

	request = testsuite.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	body, _ := testsuite.ParseBody(response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /swagger/index.html")
	assert.Containsf(t, string(body), "swagger-ui.css", "GET /swagger/index.html")
}

// Swagger UI must use the host/scheme the page was loaded from, so the spec must not hardcode them
func TestSwaggerSpecHasNoHostAndSchemes(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	var spec struct {
		Host     string   `json:"host"`
		Schemes  []string `json:"schemes"`
		BasePath string   `json:"basePath"`
	}
	assert.NoError(t, json.Unmarshal(body, &spec))
	assert.Empty(t, spec.Host)
	assert.Empty(t, spec.Schemes)
	assert.Equal(t, "/api/v1", spec.BasePath)
}
