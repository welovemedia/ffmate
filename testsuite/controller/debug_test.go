package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestDebugSet(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodPatch, "/api/v1/debug/moo", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	assert.Equal(t, http.StatusNoContent, response.StatusCode, "POST /api/v1/debug/{namespace}")
}

func TestDebugDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	request := testsuite.NewRequest(http.MethodDelete, "/api/v1/debug", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	assert.Equal(t, http.StatusNoContent, response.StatusCode, "POST /api/v1/debug")
}
