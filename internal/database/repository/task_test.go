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

func TestDeleteInactiveRefusesActiveTasks(t *testing.T) {
	server := testserver.New(t)
	repo := (&Task{DB: server.DB()}).Setup()

	for _, status := range []dto.TaskStatus{dto.PreProcessing, dto.Running, dto.PostProcessing} {
		task, err := repo.Add(&model.Task{UUID: uuid.NewString(), Status: status})
		assert.NoError(t, err)
		deleted, err := repo.DeleteInactive(task)
		assert.NoError(t, err)
		assert.False(t, deleted, "%s task must not be deleted", status)
		found, _ := repo.First(task.UUID)
		assert.NotNil(t, found, "%s task must still exist", status)
	}

	for _, status := range []dto.TaskStatus{dto.Queued, dto.DoneSuccessful, dto.DoneError, dto.DoneCanceled} {
		task, err := repo.Add(&model.Task{UUID: uuid.NewString(), Status: status})
		assert.NoError(t, err)
		deleted, err := repo.DeleteInactive(task)
		assert.NoError(t, err)
		assert.True(t, deleted, "%s task must be deleted", status)
	}
}

func TestUpdateUnlessCanceled(t *testing.T) {
	server := testserver.New(t)
	repo := (&Task{DB: server.DB()}).Setup()

	task, err := repo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.Running, OutputFile: &dto.RawResolved{Raw: "out-${UUID}.mp4"}})
	assert.NoError(t, err)

	// regular progress update of a running task: all fields (incl. json columns and labels) are written
	task.Progress = 42
	task.Remaining = 7
	task.OutputFile.Resolved = "out-1.mp4"
	task.Labels = []model.Label{{Value: "gpu"}}
	updated, err := repo.UpdateUnlessCanceled(task)
	assert.NoError(t, err)
	if assert.NotNil(t, updated) {
		assert.Equal(t, 42.0, updated.Progress)
		assert.Equal(t, "out-1.mp4", updated.OutputFile.Resolved)
		assert.Equal(t, []string{"gpu"}, []string{updated.Labels[0].Value})
	}

	// a cancel stored in the meantime must not be overwritten by a progress update
	assert.NoError(t, server.DB().Model(&model.Task{}).Where("uuid = ?", task.UUID).Update("status", dto.DoneCanceled).Error)
	task.Progress = 80
	task.Status = dto.Running
	updated, err = repo.UpdateUnlessCanceled(task)
	assert.NoError(t, err)
	assert.Nil(t, updated)
	stored, _ := repo.First(task.UUID)
	assert.Equal(t, dto.DoneCanceled, stored.Status)
	assert.Equal(t, 42.0, stored.Progress, "nothing is written for a canceled task")

	// a deleted task is neither updated nor re-inserted
	other, err := repo.Add(&model.Task{UUID: uuid.NewString(), Status: dto.Running})
	assert.NoError(t, err)
	assert.NoError(t, repo.Delete(other))
	updated, err = repo.UpdateUnlessCanceled(other)
	assert.NoError(t, err)
	assert.Nil(t, updated)
	found, _ := repo.First(other.UUID)
	assert.Nil(t, found)
}
