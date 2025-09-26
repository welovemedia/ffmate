package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
)

func TestCountAllStatusWithEmptySession(t *testing.T) {
	server := testserver.New(t)
	repo := (&Task{DB: server.DB()}).Setup()

	for _, i := range []dto.TaskStatus{dto.Queued, dto.Running, dto.DoneCanceled, dto.DoneError, dto.DoneSuccessful} {
		_, _ = repo.Add(&model.Task{
			Name:   "test",
			Status: i,
		})
	}

	q, r, ds, de, dc, err := repo.CountAllStatus()
	assert.NoError(t, err)
	assert.Equal(t, 1, q)
	assert.Equal(t, 1, r)
	assert.Equal(t, 1, ds)
	assert.Equal(t, 1, de)
	assert.Equal(t, 1, dc)
}

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

	q, err := repo.NextQueued(3)
	assert.NoError(t, err)
	assert.Nil(t, q)
}
