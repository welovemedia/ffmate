package cfg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// []string
	Set("test.key", []string{"test-value", "test-value-2"})
	assert.Equal(t, "test-value", GetStringSlice("test.key")[0], "Config GetString")
	assert.Equal(t, "test-value-2", GetStringSlice("test.key")[1], "Config GetString")
	assert.Equal(t, 2, len(GetStringSlice("test.key")), "Config GetString")

	// string
	Set("test.key", "test-value")
	assert.Equal(t, "test-value", GetString("test.key"), "Config GetString")

	// bool
	Set("test.key", true)
	assert.True(t, GetBool("test.key"), "Config GetBool")

	// int
	i := int(time.Now().UnixMilli())
	Set("test.key", i)
	assert.Equal(t, i, GetInt("test.key"), "Config GetInt")

	// uint
	d := uint(time.Now().UnixMilli())
	Set("test.key", d)
	assert.Equal(t, d, GetUint("test.key"), "Config GetUint")

	// interface
	Set("test.key", map[string]string{"foo": "bar"})
	assert.Equal(t, map[string]string{"foo": "bar"}, Get("test.key"), "Config Get")

	// Has
	Set("test.key", "test-value")
	assert.True(t, Has("test.key"), "Config Has")
	assert.False(t, Has("test.key.no"), "Config Has")

	// typed
	Set("test.key", "test-value")
	assert.Equal(t, "test-value", GetTyped[string]("test.key"), "Config GetTyped")

	// typed default
	assert.Equal(t, "test-value", GetOrDefault[string]("unknown.key", "test-value"), "Config GetOrDefault")
	assert.Equal(t, "test-value", GetOrDefault("unknown.key", "test-value"), "Config GetOrDefault")
	assert.Equal(t, 123, GetOrDefault[int]("unknown.key", 123), "Config GetOrDefault")
	assert.Equal(t, 123, GetOrDefault("unknown.key", 123), "Config GetOrDefault")
	assert.True(t, GetOrDefault[bool]("unknown.key", true), "Config GetOrDefault")
	assert.True(t, GetOrDefault("unknown.key", true), "Config GetOrDefault")
}
