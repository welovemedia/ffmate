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

var umami = &dto.Umami{
	Type: "event",
	Payload: dto.UmamiPayload{
		Hostname:  "localhost",
		Langugage: "en-US",
		Screen:    "3840x1600",
		URL:       "/ui/test",
		Referrer:  "",
		Title:     "ffmate",
		Website:   "ffmate",
	},
}

func TestUmami(t *testing.T) {
	server := testsuite.InitServer(t)

	body, _ := json.Marshal(umami)
	request := testsuite.NewRequest(http.MethodPost, "/umami", bytes.NewReader(body))
	request.Header.Add("Content-Type", "application/json")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	assert.Equal(t, http.StatusNoContent, response.StatusCode, "POST /umami")
}
