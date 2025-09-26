package exception

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInternalServerError(t *testing.T) {
	err := errors.New("something went wrong")
	httpErr := InternalServerError(err)

	assert.Equal(t, 500, httpErr.HTTPCode)
	assert.Equal(t, "001.000.0000", httpErr.Code)
	assert.Equal(t, "internal.server.error", httpErr.Error)
	assert.Equal(t, "something went wrong", httpErr.Message)
	assert.Empty(t, httpErr.Docs)
}

func TestHTTPInvalidRequest(t *testing.T) {
	httpErr := HTTPInvalidRequest()

	assert.Equal(t, 400, httpErr.HTTPCode)
	assert.Equal(t, "001.000.0002", httpErr.Code)
	assert.Equal(t, "invalid.request", httpErr.Error)
	assert.Equal(t, "invalid request", httpErr.Message)
	assert.Empty(t, httpErr.Docs)
}

func TestHTTPBadRequest(t *testing.T) {
	err := errors.New("missing field")
	docs := "https://example.com/docs"
	httpErr := HTTPBadRequest(err, docs)

	assert.Equal(t, 400, httpErr.HTTPCode)
	assert.Equal(t, "001.000.0003", httpErr.Code)
	assert.Equal(t, "bad.request", httpErr.Error)
	assert.Equal(t, "missing field", httpErr.Message)
	assert.Equal(t, docs, httpErr.Docs)
}

func TestHTTPInvalidBody(t *testing.T) {
	err := errors.New("invalid JSON")
	httpErr := HTTPInvalidBody(err)

	assert.Equal(t, 400, httpErr.HTTPCode)
	assert.Equal(t, "001.000.0005", httpErr.Code)
	assert.Equal(t, "invalid.request.body", httpErr.Error)
	assert.Equal(t, "invalid JSON", httpErr.Message)
	assert.Empty(t, httpErr.Docs)
}

func TestHTTPInvalidParam(t *testing.T) {
	paramName := "id"
	httpErr := HTTPInvalidParam(paramName)

	assert.Equal(t, 400, httpErr.HTTPCode)
	assert.Equal(t, "001.000.0006", httpErr.Code)
	assert.Equal(t, "invalid.param", httpErr.Error)
	assert.Equal(t, "invalid parameter 'id'", httpErr.Message)
	assert.Empty(t, httpErr.Docs)
}

func TestHTTPInvalidQuery(t *testing.T) {
	queryName := "page"
	httpErr := HTTPInvalidQuery(queryName)

	assert.Equal(t, 400, httpErr.HTTPCode)
	assert.Equal(t, "001.000.0007", httpErr.Code)
	assert.Equal(t, "invalid.query", httpErr.Error)
	assert.Equal(t, "invalid query 'page'", httpErr.Message)
	assert.Empty(t, httpErr.Docs)
}

func TestHTTPNotFound(t *testing.T) {
	err := errors.New("resource not found")
	docs := "https://example.com/docs"
	httpErr := HTTPNotFound(err, docs)

	assert.Equal(t, 400, httpErr.HTTPCode)
	assert.Equal(t, "001.000.0008", httpErr.Code)
	assert.Equal(t, "not.found", httpErr.Error)
	assert.Equal(t, "resource not found", httpErr.Message)
	assert.Equal(t, docs, httpErr.Docs)
}
