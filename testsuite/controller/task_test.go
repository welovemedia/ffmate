package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/internal/cfg"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/service/client"
	"github.com/welovemedia/ffmate/testsuite"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/util/testutil"
)

var newTask = &dto.NewTask{
	Name:       "Test task",
	Command:    "-y",
	Priority:   100,
	OutputFile: "/dev/null",
}

func createTask(t *testing.T, server *testutil.TestServer) *http.Response {
	body, _ := json.Marshal(newTask)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/tasks")
	return response

}

func TestTaskCreate(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createTask(t, server)
	task, _ := testsuite.ParseJsonBody[dto.Task](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/tasks")
	assert.Equal(t, task.Name, "Test task", "POST /api/v1/tasks")
	assert.Equal(t, task.Status, dto.QUEUED, "POST /api/v1/tasks")
	assert.NotEmpty(t, task.Uuid, "POST /api/v1/tasks")
}

func TestTaskList(t *testing.T) {
	server := testsuite.InitServer(t)

	createTask(t, server)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	response := server.TestRequest(request)

	tasks, _ := testsuite.ParseJsonBody[[]dto.Task](response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode, "GET /api/v1/tasks")
	assert.Equal(t, 1, len(tasks), "GET /api/v1/tasks")
	assert.Equal(t, response.Header.Get("X-Total"), "1", "GET /api/v1/tasks")
}

func TestTaskDelete(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createTask(t, server)
	task, _ := testsuite.ParseJsonBody[dto.Task](response.Body)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+task.Uuid, nil)
	response = server.TestRequest(request)
	assert.Equal(t, 204, response.StatusCode, "DELETE /api/v1/tasks")

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+task.Uuid, nil)
	response = server.TestRequest(request)
	assert.Equal(t, 400, response.StatusCode, "DELETE /api/v1/tasks")
}

func TestTaskGet(t *testing.T) {
	server := testsuite.InitServer(t)

	response := createTask(t, server)
	task, _ := testsuite.ParseJsonBody[dto.Task](response.Body)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+task.Uuid, nil)
	response = server.TestRequest(request)
	task, _ = testsuite.ParseJsonBody[dto.Task](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, task.Name, "Test task", "GET /api/v1/tasks/{uuid}")
}

func TestTaskUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	server := testsuite.InitServer(t)

	response := createTask(t, server)
	task, _ := testsuite.ParseJsonBody[dto.Task](response.Body)
	assert.Equal(t, task.Client.Identifier, "test-client", "GET /api/v1/tasks/{uuid}")

	// change client identifier
	svc, _ := server.Service(service.Client).(*client.Service)
	cfg.Set("ffmate.identifier", "test-client-changed")
	svc.UpdateClientInfo()

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+task.Uuid+"/cancel", nil)
	response = server.TestRequest(request)
	task, _ = testsuite.ParseJsonBody[dto.Task](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, task.Status, dto.DONE_CANCELED, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, task.Client.Identifier, "test-client-changed", "GET /api/v1/tasks/{uuid}")

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+task.Uuid+"/restart", nil)
	response = server.TestRequest(request)
	task, _ = testsuite.ParseJsonBody[dto.Task](response.Body)
	assert.Equal(t, 200, response.StatusCode, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, task.Status, dto.QUEUED, "GET /api/v1/tasks/{uuid}")
	assert.Equal(t, task.Client.Identifier, "test-client-changed", "GET /api/v1/tasks/{uuid}")
}

func TestTaskCreateBatch(t *testing.T) {
	server := testsuite.InitServer(t)

	// create batch
	batch := dto.NewBatch{
		Tasks: []*dto.NewTask{
			&dto.NewTask{
				Name:    "Test task 1",
				Command: "",
			},
		},
	}
	b, _ := json.Marshal(batch)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/batches", bytes.NewReader(b))
	request.Header.Set("Content-Type", "application/json")
	response := server.TestRequest(request)
	body, _ := testsuite.ParseJsonBody[dto.Batch](response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "POST /api/v1/batches")
	assert.NotEmpty(t, body.Uuid, "POST /api/v1/batches")

	// list tasks for batch
	request = httptest.NewRequest(http.MethodGet, "/api/v1/batches/"+body.Uuid, nil)
	response = server.TestRequest(request)
	body2, _ := testsuite.ParseJsonBody[dto.Batch](response.Body)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Get /api/v1/batches/{uuid}")
	assert.Equal(t, len(body2.Tasks), 1, "GET /api/v1/batches/{uuid}")
	assert.Equal(t, body2.Tasks[0].Batch, body.Uuid, "GET /api/v1/batches/{uuid}")
	assert.Equal(t, body2.Uuid, body.Uuid, "GET /api/v1/batches/{uuid}")
}
