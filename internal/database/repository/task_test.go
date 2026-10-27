package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
)

func TestCountUnfinishedByBatch(t *testing.T) {
	server := testserver.New(t)
	repo := (&Task{DB: server.DB()}).Setup()

	q, err := repo.CountUnfinishedByBatch(uuid.NewString())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), q)
}

func TestNextQueued(t *testing.T) {
	server := testserver.New(t)
	repo := (&Task{DB: server.DB()}).Setup()

	q, err := repo.NextQueued(3, cfg.GetStringSlice("ffmate.labels"))
	assert.NoError(t, err)
	assert.Nil(t, q)
}
