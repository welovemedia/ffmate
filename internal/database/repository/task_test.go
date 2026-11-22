package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
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

func TestUpdateFieldsDoesNotResurrectDeletedTask(t *testing.T) {
	server := testserver.New(t)
	repo := (&Task{DB: server.DB()}).Setup()

	task, err := repo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.Running})
	assert.NoError(t, err)

	updated, err := repo.UpdateFields(task.UUID, map[string]any{"progress": 50.0})
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, 50.0, updated.Progress)

	assert.NoError(t, repo.Delete(task))

	updated, err = repo.UpdateFields(task.UUID, map[string]any{"status": dto.DoneCanceled})
	assert.NoError(t, err)
	assert.Nil(t, updated)

	found, err := repo.First(task.UUID)
	assert.NoError(t, err)
	assert.Nil(t, found, "deleted task must not come back")

	count, err := repo.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
