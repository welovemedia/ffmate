package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
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
	OutputFile:  "/dev/null",
	PreProcessing: &dto.NewPrePostProcessing{
		SidecarPath: "/dev/null",
		ScriptPath:  "/dev/null",
	},
	PostProcessing: &dto.NewPrePostProcessing{
		SidecarPath: "/dev/null",
		ScriptPath:  "/dev/null",
	},
	Webhooks: &dto.DirectWebhooks{
		dto.NewWebhook{
			Event: dto.TaskCreated,
			URL:   "https://example.com",
		},
	},
	GlobalPresetName: "moo",
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
	defer response.Body.Close() // nolint:errcheck
	preset, _ := testsuite.ParseJSONBody[dto.Preset](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/presets")
	assert.Equal(t, "Test preset", preset.Name, "POST /api/v1/presets")
	assert.Len(t, *preset.Webhooks, 1, "POST /api/v1/presets")
	assert.NotNil(t, preset.PreProcessing, "POST /api/v1/presets")
	assert.NotNil(t, preset.PostProcessing, "POST /api/v1/presets")
	assert.False(t, preset.PreProcessing.ImportSidecar, "POST /api/v1/presets")
	assert.Equal(t, "/dev/null", preset.PreProcessing.SidecarPath, "POST /api/v1/presets")
	assert.Equal(t, "/dev/null", preset.PreProcessing.ScriptPath, "POST /api/v1/presets")
	assert.Equal(t, "/dev/null", preset.PostProcessing.SidecarPath, "POST /api/v1/presets")
	assert.Equal(t, "/dev/null", preset.PostProcessing.ScriptPath, "POST /api/v1/presets")
	assert.NotEmpty(t, preset.UUID, "POST /api/v1/presets")
}

func TestPresetList(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	defer response.Body.Close() // nolint:errcheck

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/presets", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	presets, _ := testsuite.ParseJSONBody[[]dto.Preset](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/presets")
	assert.Len(t, presets, 1, "GET /api/v1/presets")
	assert.Equal(t, "1", response.Header.Get("X-Total"), "GET /api/v1/presets")
}

func TestPresetDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	defer response.Body.Close() // nolint:errcheck
	preset, _ := testsuite.ParseJSONBody[dto.Preset](response.Body)

	request := testsuite.NewRequest(http.MethodDelete, "/api/v1/presets/"+preset.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 204, response.StatusCode, "DELETE /api/v1/presets")

	request = testsuite.NewRequest(http.MethodDelete, "/api/v1/presets/"+preset.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 400, response.StatusCode, "DELETE /api/v1/presets")
}

func TestPresetGet(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	defer response.Body.Close() // nolint:errcheck
	preset, _ := testsuite.ParseJSONBody[dto.Preset](response.Body)

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/presets/"+preset.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	preset, _ = testsuite.ParseJSONBody[dto.Preset](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/presets/{uuid}")
	assert.Equal(t, "Test preset", preset.Name, "GET /api/v1/preset/{uuid}")
}

func TestPresetUpdate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createPreset(t, server)
	defer response.Body.Close() // nolint:errcheck
	preset, _ := testsuite.ParseJSONBody[dto.Preset](response.Body)

	preset.Name = "Test Updated preset"
	preset.Labels = append(preset.Labels, "test-label-4")
	body, _ := json.Marshal(preset)
	request := testsuite.NewRequest(http.MethodPut, "/api/v1/presets/"+preset.UUID, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	preset, _ = testsuite.ParseJSONBody[dto.Preset](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/presets/{uuid}")
	assert.Contains(t, preset.Labels, "test-label-4", "POST /api/v1/presets")
	assert.Equal(t, "Test Updated preset", preset.Name, "GET /api/v1/preset/{uuid}")
	assert.Equal(t, "Test Updated preset", preset.Name, "GET /api/v1/preset/{uuid}")

	// update with invalid UUID
	preset.Name = "Test Updated preset fail"
	body, _ = json.Marshal(preset)
	request = testsuite.NewRequest(http.MethodPut, "/api/v1/presets/"+uuid.NewString(), bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	preset, _ = testsuite.ParseJSONBody[dto.Preset](response.Body)
	assert.Equal(t, 400, response.StatusCode, "GET /api/v1/presets/{uuid}")
}
