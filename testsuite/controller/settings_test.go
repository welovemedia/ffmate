package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestSettingsLoad(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/settings")
}

func TestSettingsStore(t *testing.T) {
	server := testsuite.InitServer(t)

	b, _ := json.Marshal(&dto.Settings{})
	request := testsuite.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(b))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/settings")
}
