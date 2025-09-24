package watchfolder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/internal/database/model"
)

func TestHasLabelOverlap(t *testing.T) {
	a := []model.Label{model.Label{ID: 1, Value: "test"}}
	assert.True(t, hasLabelOverlap(a, []string{"test"}), "Labels overlap")
	assert.False(t, hasLabelOverlap(a, []string{"no-test"}), "Labels overlap")
}
