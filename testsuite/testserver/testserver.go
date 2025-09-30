package testserver

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/middleware/parse"
	"goyave.dev/goyave/v5/util/testutil"
)

var c = map[string]any{
	"app": map[string]any{
		"name":    "ffmate",
		"version": "test-1.0.0",
	},
	"database": map[string]any{
		"connection": "sqlite3",
		"name":       ":memory:",
	},
	"auth": map[string]any{
		"basic": map[string]any{
			"username": "user",
			"password": "pass",
		},
	},
}

func New(t *testing.T) *testutil.TestServer {
	b, _ := json.Marshal(c)
	conf, _ := config.LoadJSON(string(b))

	if !cfg.Has("ffmate.identifier") {
		cfg.Set("ffmate.identifier", "test-client")
	}
	if !cfg.Has("ffmate.labels") {
		cfg.Set("ffmate.labels", []string{"test-label-1", "test-label-2", "test-label-3"})
	}
	if !cfg.Has("ffmate.isAuth") {
		cfg.Set("ffmate.isAuth", false)
	}
	cfg.Set("ffmate.session", uuid.NewString())
	cfg.Set("ffmate.maxConcurrentTasks", 3)
	cfg.Set("ffmate.isTray", true)
	cfg.Set("ffmate.isCluster", false)
	cfg.Set("ffmate.isFFmpeg", false)
	cfg.Set("ffmate.ffmpeg", "ffmpeg")
	cfg.Set("ffmate.debug", "")
	cfg.Set("ffmate.isDocker", false)
	cfg.Set("ffmate.isCluster", false)

	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: conf})

	// add global parsing
	server.Router().GlobalMiddleware(&parse.Middleware{
		MaxUploadSize: 10,
	})
	return server
}
