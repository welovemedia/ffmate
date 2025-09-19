package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var svc = NewService("0.0.0")

func TestCheckForUpdate(t *testing.T) {
	v, ok, err := svc.CheckForUpdate(false, true)

	assert.Nil(t, err, "Error")
	assert.True(t, ok, "Ok")
	assert.NotEqual(t, "0.0.0", v, "Version")
}
