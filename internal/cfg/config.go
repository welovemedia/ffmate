package cfg

import (
	"fmt"
	"sync"
)

var cfg = sync.Map{}

// Set stores a key/value pair safely
func Set(key string, value any) {
	cfg.Store(key, value)
}

// Get returns a raw interface
func Get(key string) any {
	if value, ok := cfg.Load(key); !ok {
		panic("config key not found: " + key)
	} else {
		return value
	}
}

func GetTyped[T any](key string) T {
	val := Get(key)
	typed, ok := val.(T)
	if !ok {
		panic(fmt.Sprintf("config key %s has wrong type: %T != %T", key, val, *new(T)))
	}
	return typed
}

func GetOrDefault[T any](key string, def T) T {
	val, ok := cfg.Load(key)
	if !ok {
		return def
	}
	typed, ok := val.(T)
	if !ok {
		panic(fmt.Sprintf("config key %s has wrong type: %T != %T", key, val, *new(T)))
	}
	return typed
}

func Has(key string) bool {
	_, ok := cfg.Load(key)
	return ok
}

// Typed helpers (like viper)
func GetStringSlice(key string) []string {
	val := Get(key)
	typed, ok := val.([]string)
	if !ok {
		panic("config key is not []string: " + key)
	}
	return typed
}

func GetString(key string) string {
	val := Get(key)
	s, ok := val.(string)
	if !ok {
		panic("config key is not a string: " + key)
	}
	return s
}

func GetBool(key string) bool {
	val := Get(key)
	b, ok := val.(bool)
	if !ok {
		panic("config key is not a boolean: " + key)
	}
	return b
}

func GetInt(key string) int {
	val := Get(key)
	i, ok := val.(int)
	if !ok {
		panic("config key is not an int: " + key)
	}
	return i
}

func GetUint(key string) uint {
	val := Get(key)
	u, ok := val.(uint)
	if !ok {
		panic("config key is not a uint: " + key)
	}
	return u
}
