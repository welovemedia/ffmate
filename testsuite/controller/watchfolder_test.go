package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/util/testutil"
)

var newWatchfolder = &dto.Watchfolder{
	Name:        "Test watchfolder",
	Description: "Test desc",

	Path:   "/dev/null",
	Preset: "123",

	Interval:     5,
	GrowthChecks: 1,

	Suspended: false,

	Filter: &dto.WatchfolderFilter{
		Extensions: &dto.WatchfolderFilterExtensions{
			Include: []string{"mov", "mp4"},
			Exclude: []string{"mov", "mp4"},
		},
	},
}

func init() {
	cfg.Set("ffmate.isCluster", false)
}

func createWatchfolder(t *testing.T, server *testutil.TestServer) *http.Response {
	body, _ := json.Marshal(newWatchfolder)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/watchfolders", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/watchfolders")
	return response
}

func TestWatchfolderCreate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWatchfolder(t, server)
	defer response.Body.Close() // nolint:errcheck
	watchfolder, _ := testsuite.ParseJSONBody[dto.Watchfolder](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/watchfolders")
	assert.Equal(t, "Test watchfolder", watchfolder.Name, "POST /api/v1/watchfolders")
	assert.NotEmpty(t, watchfolder.UUID, "POST /api/v1/watchfolders")
}

func TestWatchfolderList(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWatchfolder(t, server)
	defer response.Body.Close() // nolint:errcheck

	request := httptest.NewRequest(http.MethodGet, "/api/v1/watchfolders", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	watchfolder, _ := testsuite.ParseJSONBody[[]dto.Watchfolder](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/watchfolders")
	assert.Len(t, watchfolder, 1, "GET /api/v1/watchfolders")
	assert.Equal(t, "1", response.Header.Get("X-Total"), "GET /api/v1/watchfolders")
}

func TestWatchfolderDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWatchfolder(t, server)
	defer response.Body.Close() // nolint:errcheck
	watchfolder, _ := testsuite.ParseJSONBody[dto.Watchfolder](response.Body)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/watchfolders/"+watchfolder.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 204, response.StatusCode, "DELETE /api/v1/watchfolders")

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/watchfolders/"+watchfolder.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 400, response.StatusCode, "DELETE /api/v1/watchfolders")
}

func TestWatchfolderGet(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWatchfolder(t, server)
	defer response.Body.Close() // nolint:errcheck
	watchfolder, _ := testsuite.ParseJSONBody[dto.Watchfolder](response.Body)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/watchfolders/"+watchfolder.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	watchfolder, _ = testsuite.ParseJSONBody[dto.Watchfolder](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/watchfolders/{uuid}")
	assert.Equal(t, "Test watchfolder", watchfolder.Name, "GET /api/v1/watchfolders/{uuid}")

	request = httptest.NewRequest(http.MethodGet, "/api/v1/watchfolders/"+uuid.NewString(), nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 400, response.StatusCode, "GET /api/v1/watchfolders/{uuid}")
}

func TestWatchfolderUpdate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createWatchfolder(t, server)
	defer response.Body.Close() // nolint:errcheck
	watchfolder, _ := testsuite.ParseJSONBody[dto.Watchfolder](response.Body)

	watchfolder.Name = "Test Updated watchfolder"
	body, _ := json.Marshal(watchfolder)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/watchfolders/"+watchfolder.UUID, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	watchfolder, _ = testsuite.ParseJSONBody[dto.Watchfolder](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/watchfolders/{uuid}")
	assert.Equal(t, "Test Updated watchfolder", watchfolder.Name, "GET /api/v1/watchfolders/{uuid}")
}
