package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
)

func TestSetupConfig(t *testing.T) {
	viper.Set("database", "postgresql://localhost:5432/testdb")
	viper.Set("ffmpeg", "/usr/bin/ffmpeg")
	viper.Set("debug", "info:*")
	viper.Set("maxConcurrentTasks", 5)
	viper.Set("tray", true)
	viper.Set("noUI", false)
	viper.Set("identifier", "")
	viper.Set("sendTelemetry", true)

	// Remove /.dockerenv to simulate non-Docker environment
	_ = os.Remove("/.dockerenv")

	setupConfig()

	assert.Equal(t, "/usr/bin/ffmpeg", cfg.GetString("ffmate.ffmpeg"))
	assert.Equal(t, "info:*", cfg.GetString("ffmate.debug"))
	assert.Equal(t, 5, cfg.GetInt("ffmate.maxConcurrentTasks"))
	assert.Equal(t, "postgresql://localhost:5432/testdb", cfg.GetString("ffmate.database"))
	assert.True(t, cfg.GetBool("ffmate.isCluster"))
	assert.True(t, cfg.GetBool("ffmate.isTray"))
	assert.True(t, cfg.GetBool("ffmate.isUI"))
	assert.NotEmpty(t, cfg.GetString("ffmate.session"))
	assert.NotEmpty(t, cfg.GetString("ffmate.identifier"))
	assert.True(t, cfg.GetBool("ffmate.telemetry.send"))
	assert.Equal(t, "https://telemetry.ffmate.io", cfg.GetString("ffmate.telemetry.url"))
}

func TestSetupGoyaveConfig_Postgres(t *testing.T) {
	viper.Set("database", "postgresql://user:pass@localhost:5432/testdb?sslmode=disable")
	viper.Set("app.name", "ffmate")
	viper.Set("app.version", "1.2.3")
	viper.Set("port", 8080)

	cfg := setupGoyaveConfig()

	assert.Equal(t, "postgres", cfg.Get("database.connection"))
	assert.Equal(t, "localhost:5432", cfg.Get("database.host"))
	assert.Equal(t, "testdb", cfg.Get("database.name"))
	assert.Equal(t, "user", cfg.Get("database.username"))
	assert.Equal(t, "pass", cfg.Get("database.password"))
	assert.Equal(t, "?sslmode=disable", cfg.Get("database.options"))
}

func TestSetupGoyaveConfig_SQLite(t *testing.T) {
	viper.Set("database", "~/test.sqlite")
	viper.Set("app.name", "ffmate")
	viper.Set("app.version", "1.2.3")
	viper.Set("port", 8080)

	cfg := setupGoyaveConfig()

	assert.Equal(t, "sqlite3", cfg.Get("database.connection"))
}
