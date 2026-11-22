package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
)

func TestFirstAndLockOnlyClaimsWhenEligible(t *testing.T) {
	server := testserver.New(t)
	repo := (&Watchfolder{DB: server.DB()}).Setup()

	w, err := repo.Save(&model.Watchfolder{UUID: uuid.NewString(), Interval: 60})
	assert.NoError(t, err)

	// an ineligible instance (e.g. labels mismatch) must not claim the interval
	wf, locked, err := repo.FirstAndLock(w.UUID, func(*model.Watchfolder) bool { return false })
	assert.NoError(t, err)
	assert.NotNil(t, wf)
	assert.False(t, locked)

	// an eligible instance can still claim it
	_, locked, err = repo.FirstAndLock(w.UUID, func(*model.Watchfolder) bool { return true })
	assert.NoError(t, err)
	assert.False(t, locked, "eligible instance must get the interval")

	// and afterwards it is locked for everyone else
	_, locked, err = repo.FirstAndLock(w.UUID, func(*model.Watchfolder) bool { return true })
	assert.NoError(t, err)
	assert.True(t, locked)
}
