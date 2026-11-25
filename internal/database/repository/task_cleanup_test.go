package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
	"gorm.io/gorm"
)

// seedClientTasks creates a client and one task per status for it, returns uuid by status
func seedClientTasks(t *testing.T, db *gorm.DB, identifier string, lastSeen int64) map[dto.TaskStatus]string {
	t.Helper()
	assert.NoError(t, db.Create(&model.Client{Identifier: identifier, LastSeen: lastSeen}).Error)
	repo := &Task{DB: db}
	uuids := map[dto.TaskStatus]string{}
	for _, status := range []dto.TaskStatus{dto.Queued, dto.PreProcessing, dto.Running, dto.PostProcessing, dto.DoneSuccessful, dto.DoneCanceled} {
		task, err := repo.Add(&model.Task{UUID: uuid.NewString(), Status: status, ClientIdentifier: identifier, Labels: []model.Label{{Value: "gpu"}}})
		assert.NoError(t, err)
		uuids[status] = task.UUID
	}
	return uuids
}

func assertCleanedUp(t *testing.T, repo *Task, failed []model.Task, uuids map[dto.TaskStatus]string) {
	t.Helper()
	assert.Len(t, failed, 3, "pre-processing, running and post-processing tasks are failed")
	for _, task := range failed {
		assert.Equal(t, dto.DoneError, task.Status)
		assert.Equal(t, "client disconnected during execution", task.Error)
		assert.Equal(t, []string{"gpu"}, []string{task.Labels[0].Value}, "labels are loaded for webhooks")
	}
	for _, status := range []dto.TaskStatus{dto.PreProcessing, dto.Running, dto.PostProcessing} {
		task, _ := repo.First(uuids[status])
		assert.Equal(t, dto.DoneError, task.Status, "%s task must be failed", status)
	}
	for _, status := range []dto.TaskStatus{dto.Queued, dto.DoneSuccessful, dto.DoneCanceled} {
		task, _ := repo.First(uuids[status])
		assert.Equal(t, status, task.Status, "%s task must be untouched", status)
	}
}

func TestFailRunningTasksForStartingClient(t *testing.T) {
	server := testserver.New(t)
	(&Client{DB: server.DB()}).Setup()
	repo := (&Task{DB: server.DB()}).Setup()

	uuids := seedClientTasks(t, server.DB(), "crashed", time.Now().UnixMilli())
	other := seedClientTasks(t, server.DB(), "other", time.Now().UnixMilli())

	failed, err := repo.FailRunningTasksForStartingClient("crashed")
	assert.NoError(t, err)
	assertCleanedUp(t, repo, failed, uuids)

	task, _ := repo.First(other[dto.Running])
	assert.Equal(t, dto.Running, task.Status, "tasks of other clients must be untouched")

	failed, err = repo.FailRunningTasksForStartingClient("crashed")
	assert.NoError(t, err)
	assert.Empty(t, failed, "already failed tasks are not returned again")
}

func TestFailRunningTasksForOfflineClients(t *testing.T) {
	server := testserver.New(t)
	(&Client{DB: server.DB()}).Setup()
	repo := (&Task{DB: server.DB()}).Setup()

	offline := seedClientTasks(t, server.DB(), "offline", time.Now().Add(-2*time.Minute).UnixMilli())
	online := seedClientTasks(t, server.DB(), "online", time.Now().UnixMilli())

	failed, err := repo.FailRunningTasksForOfflineClients()
	assert.NoError(t, err)
	assertCleanedUp(t, repo, failed, offline)

	task, _ := repo.First(online[dto.Running])
	assert.Equal(t, dto.Running, task.Status, "tasks of online clients must be untouched")
}
