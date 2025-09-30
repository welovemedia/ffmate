package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/client"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/util/testutil"
)

var newTask = &dto.NewTask{
	Name:       "Test task",
	Command:    "-y",
	Priority:   100,
	Labels:     []string{"test-label-1", "test-label-2", "test-label-3"},
	OutputFile: "/dev/null",
	Metadata: &dto.MetadataMap{
		"foo": "bar",
	},
}

func createTask(t *testing.T, server *testutil.TestServer) *http.Response {
	body, _ := json.Marshal(newTask)
	request := testsuite.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/tasks")
	return response
}

func TestTaskCreate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createTask(t, server)
	defer response.Body.Close() // nolint:errcheck
	task, _ := testsuite.ParseJSONBody[dto.Task](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/tasks")
	assert.Equal(t, task.Name, "Test task", "POST /api/v1/tasks")
	assert.Contains(t, task.Labels, "test-label-1", "POST /api/v1/tasks")
	assert.Contains(t, task.Labels, "test-label-2", "POST /api/v1/tasks")
	assert.Contains(t, task.Labels, "test-label-3", "POST /api/v1/tasks")
	assert.NotContains(t, task.Labels, "test-label-0", "POST /api/v1/tasks")
	assert.Equal(t, task.Status, dto.Queued, "POST /api/v1/tasks")
	assert.Equal(t, "Test task", task.Name, "POST /api/v1/tasks")
	assert.Equal(t, dto.Queued, task.Status, "POST /api/v1/tasks")
	assert.NotNil(t, task.Metadata, "POST /api/v1/tasks")
	assert.NotEmpty(t, task.UUID, "POST /api/v1/tasks")
}

func TestTaskList(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createTask(t, server)
	defer response.Body.Close() // nolint:errcheck

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck

	tasks, _ := testsuite.ParseJSONBody[[]dto.Task](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/tasks")
	assert.Len(t, tasks, 1, "GET /api/v1/tasks")
	assert.Equal(t, "1", response.Header.Get("X-Total"), "GET /api/v1/tasks")
}

func TestTaskDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createTask(t, server)
	defer response.Body.Close() // nolint:errcheck
	task, _ := testsuite.ParseJSONBody[dto.Task](response.Body)

	request := testsuite.NewRequest(http.MethodDelete, "/api/v1/tasks/"+task.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 204, response.StatusCode, "DELETE /api/v1/tasks")

	request = testsuite.NewRequest(http.MethodDelete, "/api/v1/tasks/"+task.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	assert.Equal(t, 400, response.StatusCode, "DELETE /api/v1/tasks")
}

func TestTaskGet(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createTask(t, server)
	defer response.Body.Close() // nolint:errcheck
	task, _ := testsuite.ParseJSONBody[dto.Task](response.Body)

	request := testsuite.NewRequest(http.MethodGet, "/api/v1/tasks/"+task.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	task, _ = testsuite.ParseJSONBody[dto.Task](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, "Test task", task.Name, "GET /api/v1/tasks/{uuid}")
}

func TestTaskUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	server := testsuite.InitServer(t)

	response := createTask(t, server)
	defer response.Body.Close() // nolint:errcheck
	task, _ := testsuite.ParseJSONBody[dto.Task](response.Body)
	assert.Equal(t, "test-client", task.Client.Identifier, "GET /api/v1/tasks/{uuid}")

	// change client identifier
	svc, _ := server.Service(service.Client).(*client.Service)
	cfg.Set("ffmate.identifier", "test-client-changed")

	svc.UpdateClientInfo(false)

	request := testsuite.NewRequest(http.MethodPatch, "/api/v1/tasks/"+task.UUID+"/cancel", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	task, _ = testsuite.ParseJSONBody[dto.Task](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, dto.DoneCanceled, task.Status, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, "test-client-changed", task.Client.Identifier, "GET /api/v1/tasks/{uuid}")

	request = testsuite.NewRequest(http.MethodPatch, "/api/v1/tasks/"+task.UUID+"/restart", nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	task, _ = testsuite.ParseJSONBody[dto.Task](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, dto.Queued, task.Status, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, "test-client-changed", task.Client.Identifier, "GET /api/v1/tasks/{uuid}")
}

func TestTaskCreateBatch(t *testing.T) {
	server := testsuite.InitServer(t)

	// create batch
	batch := dto.NewBatch{
		Tasks: []*dto.NewTask{
			{
				Name:    "Test task 1",
				Command: "",
			},
		},
	}
	b, _ := json.Marshal(batch)
	request := testsuite.NewRequest(http.MethodPost, "/api/v1/batches", bytes.NewReader(b))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	body, _ := testsuite.ParseJSONBody[dto.Batch](response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/batches")
	assert.NotEmpty(t, body.UUID, "POST /api/v1/batches")

	// list tasks for batch
	request = testsuite.NewRequest(http.MethodGet, "/api/v1/batches/"+body.UUID, nil)
	response = server.TestRequest(request)
	defer response.Body.Close() // nolint:errcheck
	body2, _ := testsuite.ParseJSONBody[dto.Batch](response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Get /api/v1/batches/{uuid}")
	assert.Len(t, body2.Tasks, 1, "GET /api/v1/batches/{uuid}")
	assert.Equal(t, body2.Tasks[0].Batch, body.UUID, "GET /api/v1/batches/{uuid}")
	assert.Equal(t, body2.UUID, body.UUID, "GET /api/v1/batches/{uuid}")
}

func TestTaskNextFromQueue(t *testing.T) {
	// tasks by label
	server := testsuite.InitServer(t)

	createTask(t, server)
	createTask(t, server)
	createTask(t, server)

	taskRepo := (&repository.Task{DB: server.DB()}).Setup()
	tasks, err := taskRepo.NextQueued(3, cfg.GetStringSlice("ffmate.labels"))
	assert.NotNil(t, tasks, "Find next tasks by labels")
	assert.NoError(t, err, tasks, "Find next tasks by labels")
	assert.Equal(t, 3, len(*tasks), "Find next tasks by labels")

	// no matching tasks
	cfg.Set("ffmate.labels", []string{"no-labels"})
	server = testsuite.InitServer(t)

	createTask(t, server)
	createTask(t, server)
	createTask(t, server)

	taskRepo = (&repository.Task{DB: server.DB()}).Setup()
	tasks, err = taskRepo.NextQueued(3, cfg.GetStringSlice("ffmate.labels"))
	assert.Nil(t, tasks, "Find next tasks by labels")

	// 1/3 matching tasks
	createTask(t, server)
	createTask(t, server)

	b := *newTask
	b.Labels = append(b.Labels, "no-labels")
	body, _ := json.Marshal(b)
	request := testsuite.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/tasks")

	taskRepo = (&repository.Task{DB: server.DB()}).Setup()
	tasks, err = taskRepo.NextQueued(3, cfg.GetStringSlice("ffmate.labels"))
	assert.NotNil(t, tasks, "Find next tasks by labels")
	assert.NoError(t, err, tasks, "Find next tasks by labels")
	assert.Equal(t, 1, len(*tasks), "Find next tasks by labels")
}
