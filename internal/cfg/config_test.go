package cfg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// string
	Set("test.key", "test-value")
	assert.Equal(t, "test-value", GetString("test.key"), "Config GetString")

	// bool
	Set("test.key", true)
	assert.Equal(t, true, GetBool("test.key"), "Config GetBool")

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
	assert.Equal(t, true, Has("test.key"), "Config Has")
	assert.Equal(t, false, Has("test.key.no"), "Config Has")

	// typed
	Set("test.key", "test-value")
	assert.Equal(t, "test-value", GetTyped[string]("test.key"), "Config GetTyped")

	// typed default
	assert.Equal(t, "test-value", GetOrDefault[string]("unknown.key", "test-value"), "Config GetOrDefault")
	assert.Equal(t, "test-value", GetOrDefault("unknown.key", "test-value"), "Config GetOrDefault")
	assert.Equal(t, 123, GetOrDefault[int]("unknown.key", 123), "Config GetOrDefault")
	assert.Equal(t, 123, GetOrDefault("unknown.key", 123), "Config GetOrDefault")
	assert.Equal(t, true, GetOrDefault[bool]("unknown.key", true), "Config GetOrDefault")
	assert.Equal(t, true, GetOrDefault("unknown.key", true), "Config GetOrDefault")
}
