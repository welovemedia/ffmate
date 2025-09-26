package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/util/testutil"
)

var newPreset = &dto.NewPreset{
	Name:        "Test preset",
	Description: "Test desc",
	Command:     "-y",
	Priority:    100,
	Labels:      []string{"test-label-1", "test-label-2", "test-label-3"},
	OutputFile:  "/dev/null",
}

func createPreset(t *testing.T, server *testutil.TestServer) *http.Response {
	body, _ := json.Marshal(newPreset)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/presets", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/presets")
	return response

}

func TestPresetCreate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	preset, _ := testsuite.ParseJsonBody[dto.Preset](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/presets")
	assert.Equal(t, preset.Name, "Test preset", "POST /api/v1/presets")
	assert.Contains(t, preset.Labels, "test-label-1", "POST /api/v1/presets")
	assert.Contains(t, preset.Labels, "test-label-2", "POST /api/v1/presets")
	assert.Contains(t, preset.Labels, "test-label-3", "POST /api/v1/presets")
	assert.NotContains(t, preset.Labels, "test-label-0", "POST /api/v1/presets")
	assert.NotEmpty(t, preset.Uuid, "POST /api/v1/presets")
}

func TestPresetList(t *testing.T) {
	server := testsuite.InitServer(t)

	createPreset(t, server)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/presets", nil)
	response := server.TestRequest(request)

	presets, _ := testsuite.ParseJsonBody[[]dto.Preset](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/presets")
	assert.Equal(t, 1, len(presets), "GET /api/v1/presets")
	assert.Equal(t, response.Header.Get("X-Total"), "1", "GET /api/v1/presets")
}

func TestPresetDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	preset, _ := testsuite.ParseJsonBody[dto.Preset](response.Body)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/presets/"+preset.Uuid, nil)
	response = server.TestRequest(request)
	assert.Equal(t, 204, response.StatusCode, "DELETE /api/v1/presets")

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/presets/"+preset.Uuid, nil)
	response = server.TestRequest(request)
	assert.Equal(t, 400, response.StatusCode, "DELETE /api/v1/presets")
}

func TestPresetGet(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	preset, _ := testsuite.ParseJsonBody[dto.Preset](response.Body)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/presets/"+preset.Uuid, nil)
	response = server.TestRequest(request)
	preset, _ = testsuite.ParseJsonBody[dto.Preset](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/presets/{uuid}")
	assert.Equal(t, preset.Name, "Test preset", "GET /api/v1/preset/{uuid}")
}

func TestPresetUpdate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	preset, _ := testsuite.ParseJsonBody[dto.Preset](response.Body)

	preset.Name = "Test Updated preset"
	preset.Labels = append(preset.Labels, "test-label-4")
	body, _ := json.Marshal(preset)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/presets/"+preset.Uuid, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response = server.TestRequest(request)
	preset, _ = testsuite.ParseJsonBody[dto.Preset](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/presets/{uuid}")
	assert.Contains(t, preset.Labels, "test-label-4", "POST /api/v1/presets")
	assert.Equal(t, preset.Name, "Test Updated preset", "GET /api/v1/preset/{uuid}")
}
