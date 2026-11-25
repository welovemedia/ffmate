package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/middleware"
	"github.com/welovemedia/ffmate/v2/testsuite"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/testutil"

	_ "github.com/welovemedia/ffmate/v2/internal/dialect"
)

func initAuthServer(t *testing.T) *testutil.TestServer {
	cfg.Set("ffmate.isAuth", true)
	t.Cleanup(func() { cfg.Set("ffmate.isAuth", false) })
	ok := func(response *goyave.Response, _ *goyave.Request) { response.Status(http.StatusNoContent) }
	return testsuite.InitServer(t, func(router *goyave.Router) {
		testRouter := router.Subrouter("/test")
		testRouter.Get("/read", ok).SetMeta(middleware.MetaAccess, middleware.AccessRead)
		testRouter.Post("/write", ok).SetMeta(middleware.MetaAccess, middleware.AccessWrite)
		testRouter.Get("/none", ok)
	})
}

func createAPIKey(t *testing.T, server *testutil.TestServer, access dto.APIKeyAccess) *dto.CreatedAPIKey {
	body, _ := json.Marshal(&dto.NewAPIKey{Name: "test", Access: access})
	request := testsuite.NewRequest(http.MethodPost, "/api/v1/apikeys", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/apikeys")
	apiKey, err := testsuite.ParseJSONBody[dto.CreatedAPIKey](response.Body)
	assert.NoError(t, err)
	return &apiKey
}

func requestWithAPIKey(server *testutil.TestServer, method, target, key string) int {
	request := testsuite.NewRequest(method, target, nil)
	request.Header.Del("Authorization")
	request.Header.Set(middleware.HeaderAPIKey, key)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	return response.StatusCode
}

func TestAPIKeyRoutesDisabledWithoutAuth(t *testing.T) {
	server := testsuite.InitServer(t)

	cases := []struct{ method, target string }{
		{http.MethodGet, "/api/v1/apikeys"},
		{http.MethodPost, "/api/v1/apikeys"},
		{http.MethodGet, "/api/v1/apikeys/some-uuid"},
		{http.MethodDelete, "/api/v1/apikeys/some-uuid"},
		{http.MethodPost, "/api/v1/apikeys/some-uuid/rotate"},
	}
	for _, c := range cases {
		request := testsuite.NewRequest(c.method, c.target, nil)
		response := server.TestRequest(request)
		assert.Equal(t, http.StatusPreconditionFailed, response.StatusCode, c.method+" "+c.target)
		_ = response.Body.Close()
	}
}

func TestAPIKeyCreate(t *testing.T) {
	server := initAuthServer(t)
	apiKey := createAPIKey(t, server, dto.APIKeyAccessRead)

	assert.NotEmpty(t, apiKey.UUID)
	assert.Equal(t, "test", apiKey.Name)
	assert.Equal(t, dto.APIKeyAccessRead, apiKey.Access)
	assert.Contains(t, apiKey.Key, "ffm_")
	assert.Equal(t, apiKey.Key[:len(apiKey.Prefix)], apiKey.Prefix)
}

func TestAPIKeyCreateInvalidAccess(t *testing.T) {
	server := initAuthServer(t)
	body, _ := json.Marshal(map[string]string{"name": "test", "access": "admin"})
	request := testsuite.NewRequest(http.MethodPost, "/api/v1/apikeys", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
}

func TestAPIKeyListGetDelete(t *testing.T) {
	server := initAuthServer(t)
	apiKey := createAPIKey(t, server, dto.APIKeyAccessWrite)

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/apikeys", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "1", response.Header.Get("X-Total"))
	list, _ := testsuite.ParseJSONBody[[]dto.APIKey](response.Body)
	assert.Len(t, list, 1)

	request = testsuite.NewRequest(http.MethodGet, "/api/v1/apikeys/"+apiKey.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode)

	request = testsuite.NewRequest(http.MethodDelete, "/api/v1/apikeys/"+apiKey.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusNoContent, response.StatusCode)

	// revoked key is rejected
	assert.Equal(t, http.StatusUnauthorized, requestWithAPIKey(server, http.MethodGet, "/test/read", apiKey.Key))
}

func TestAPIKeyAccess(t *testing.T) {
	server := initAuthServer(t)
	readKey := createAPIKey(t, server, dto.APIKeyAccessRead).Key
	writeKey := createAPIKey(t, server, dto.APIKeyAccessWrite).Key

	cases := []struct {
		name   string
		method string
		target string
		key    string
		status int
	}{
		{"read key on read route", http.MethodGet, "/test/read", readKey, http.StatusNoContent},
		{"read key on write route", http.MethodPost, "/test/write", readKey, http.StatusForbidden},
		{"read key on route without meta", http.MethodGet, "/test/none", readKey, http.StatusNoContent},
		{"write key on read route", http.MethodGet, "/test/read", writeKey, http.StatusNoContent},
		{"write key on write route", http.MethodPost, "/test/write", writeKey, http.StatusNoContent},
		{"write key on route without meta", http.MethodGet, "/test/none", writeKey, http.StatusNoContent},
		{"write key cannot manage api keys", http.MethodGet, "/api/v1/apikeys", writeKey, http.StatusForbidden},
		{"invalid key", http.MethodGet, "/test/read", "ffm_invalid", http.StatusUnauthorized},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.status, requestWithAPIKey(server, c.method, c.target, c.key))
		})
	}
}

func TestAPIKeyBasicAuthIgnoresAccessMeta(t *testing.T) {
	server := initAuthServer(t)
	request := testsuite.NewRequest(http.MethodPost, "/test/write", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestAPIKeyRotate(t *testing.T) {
	server := initAuthServer(t)
	apiKey := createAPIKey(t, server, dto.APIKeyAccessWrite)

	request := testsuite.NewRequest(http.MethodPost, "/api/v1/apikeys/"+apiKey.UUID+"/rotate", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusOK, response.StatusCode)
	rotated, err := testsuite.ParseJSONBody[dto.CreatedAPIKey](response.Body)
	assert.NoError(t, err)

	// identity and access are kept, the secret is replaced
	assert.Equal(t, apiKey.UUID, rotated.UUID)
	assert.Equal(t, apiKey.Name, rotated.Name)
	assert.Equal(t, apiKey.Access, rotated.Access)
	assert.Contains(t, rotated.Key, "ffm_")
	assert.NotEqual(t, apiKey.Key, rotated.Key)
	assert.Equal(t, rotated.Key[:len(rotated.Prefix)], rotated.Prefix)

	// old key is rejected, new key is accepted
	assert.Equal(t, http.StatusUnauthorized, requestWithAPIKey(server, http.MethodGet, "/test/read", apiKey.Key))
	assert.Equal(t, http.StatusNoContent, requestWithAPIKey(server, http.MethodPost, "/test/write", rotated.Key))
}

func TestAPIKeyRotateUnknown(t *testing.T) {
	server := initAuthServer(t)
	request := testsuite.NewRequest(http.MethodPost, "/api/v1/apikeys/unknown-uuid/rotate", nil)
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestAPIKeyRotateWithAPIKeyForbidden(t *testing.T) {
	server := initAuthServer(t)
	apiKey := createAPIKey(t, server, dto.APIKeyAccessWrite)
	assert.Equal(t, http.StatusForbidden, requestWithAPIKey(server, http.MethodPost, "/api/v1/apikeys/"+apiKey.UUID+"/rotate", apiKey.Key))
}
