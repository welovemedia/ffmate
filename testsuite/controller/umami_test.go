package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

var umami = &dto.Umami{
	Type: "event",
	Payload: dto.UmamiPayload{
		Hostname:  "localhost",
		Langugage: "en-US",
		Screen:    "3840x1600",
		Url:       "/ui/test",
		Referrer:  "",
		Title:     "ffmate",
		Website:   "ffmate",
	},
}

func TestUmami(t *testing.T) {
	server := testsuite.InitServer(t)

	body, _ := json.Marshal(umami)
	request := httptest.NewRequest(http.MethodPost, "/umami", bytes.NewReader(body))
	response := server.TestRequest(request)

	assert.Equal(t, http.StatusNoContent, response.StatusCode, "POST /umami")
}
