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
func GetString(key string) string {
	if v, ok := Get(key).(string); ok {
		return v
	}
	panic("config key is not string: " + key)
}

func GetBool(key string) bool {
	if v, ok := Get(key).(bool); ok {
		return v
	}
	panic("config key is not boolean: " + key)
}

func GetUint(key string) uint {
	if v, ok := Get(key).(uint); ok {
		return v
	}
	panic("config key is not uint: " + key)
}

func GetInt(key string) int {
	if v, ok := Get(key).(int); ok {
		return v
	}
	panic("config key is not int: " + key)
}
