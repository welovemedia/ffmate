package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBroadcastLogger(t *testing.T) {
	RegisterBroadcastLogger(func(msg []byte) {
		assert.Contains(t, string(msg), "test-message", "Msg")
		assert.Contains(t, string(msg), "info:test", "Msg")
	})

	Test.Info("test-message")
}
