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
	assert.Len(t, GetStringSlice("test.key"), 2, "Config GetString")

	// string
	Set("test.key", "test-value")
	Set("test.key.wrong.type", 1)

	assert.Equal(t, "test-value", GetString("test.key"), "Config GetString")
	assert.Panics(t, func() { GetString("test.key.wrong.type") }, "Config GetString wrong type")
	assert.Panics(t, func() { GetString("does.not.exist") }, "Config GetString does not exist")
}

func TestConfig_GetBool(t *testing.T) {
	Set("test.key", true)
	Set("test.key.wrong.type", "moo")

	assert.True(t, GetBool("test.key"), "Config GetBool")
	assert.Panics(t, func() { GetBool("test.key.wrong.type") }, "Config GetBool wrong type")
	assert.Panics(t, func() { GetBool("does.not.exist") }, "Config GetBool does not exist")
}

func TestConfig_GetInt(t *testing.T) {
	i := int(time.Now().UnixMilli())
	Set("test.key", i)
	Set("test.key.wrong.type", "moo")

	assert.Equal(t, i, GetInt("test.key"), "Config GetInt")
	assert.Panics(t, func() { GetInt("test.key.wrong.type") }, "Config GetInt wrong type")
	assert.Panics(t, func() { GetInt("does.not.exist") }, "Config GetInt does not exist")
}

func TestConfig_GetUint(t *testing.T) {
	d := uint(time.Now().UnixMilli())
	Set("test.key", d)
	Set("test.key.wrong.type", "moo")

	assert.Equal(t, d, GetUint("test.key"), "Config GetUint")
	assert.Panics(t, func() { GetUint("test.key.wrong.type") }, "Config GetUint wrong type")
	assert.Panics(t, func() { GetUint("does.not.exist") }, "Config GetUint does not exist")
}

func TestConfig_GetInterface(t *testing.T) {
	Set("test.key", map[string]string{"foo": "bar"})
	assert.Equal(t, map[string]string{"foo": "bar"}, Get("test.key"), "Config Get")
}

func TestConfig_Has(t *testing.T) {
	Set("test.key", "test-value")

	assert.True(t, Has("test.key"), "Config Has")
	assert.False(t, Has("test.key.no"), "Config Has")
}

func TestConfig_GetTyped(t *testing.T) {
	Set("test.key", "test-value")

	assert.Equal(t, "test-value", GetTyped[string]("test.key"), "Config GetTyped")
	assert.Panics(t, func() { GetTyped[int]("test.key") }, "Config wrong type")
	assert.Panics(t, func() { GetTyped[int]("does.not.exist") }, "Config does not exist")
}

func TestConfig_GetOrDefault(t *testing.T) {
	Set("known.key", "test-value")

	assert.Equal(t, "test-value", GetOrDefault[string]("known.key", "test-value"), "Config GetOrDefault existing")
	assert.Equal(t, "test-value", GetOrDefault[string]("unknown.key", "test-value"), "Config GetOrDefault string default")
	assert.Equal(t, "test-value", GetOrDefault("unknown.key", "test-value"), "Config GetOrDefault interface")
	assert.Equal(t, 123, GetOrDefault[int]("unknown.key", 123), "Config GetOrDefault int default")
	assert.Equal(t, 123, GetOrDefault("unknown.key", 123), "Config GetOrDefault int interface")
	assert.True(t, GetOrDefault[bool]("unknown.key", true), "Config GetOrDefault bool default")
	assert.True(t, GetOrDefault("unknown.key", true), "Config GetOrDefault bool interface")
}

func TestConfig_GetOrDefault_WrongType(t *testing.T) {
	cfg.Store("wrong.type", "not-a-uint")

	assert.PanicsWithValue(t,
		"config key wrong.type has wrong type: string != uint",
		func() {
			GetOrDefault[uint]("wrong.type", 0)
		},
	)
}
