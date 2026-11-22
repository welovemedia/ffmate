package controller

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/testsuite"

	_ "github.com/welovemedia/ffmate/v2/internal/dialect"
)

func TestRetriesAndPriorityValidationMessages(t *testing.T) {
	server := testsuite.InitServer(t)

	post := func(path, body string) (int, string) {
		request := testsuite.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
		request.Header.Set("Content-Type", "application/json")
		response := server.TestRequest(request)
		defer response.Body.Close() // nolint:errcheck
		b, _ := io.ReadAll(response.Body)
		return response.StatusCode, string(b)
	}

	for _, path := range []string{"/api/v1/tasks", "/api/v1/presets"} {
		for _, field := range []string{"retries", "priority"} {
			body := `{"name":"t","command":"-version","` + field + `":-1}`
			status, resp := post(path, body)
			assert.Equal(t, http.StatusUnprocessableEntity, status, "%s %s=-1", path, field)
			assert.Contains(t, resp, "must be at least 0", "%s %s=-1", path, field)
			assert.NotContains(t, resp, "positive integer", "%s %s=-1", path, field)

			status, _ = post(path, `{"name":"t","command":"-version","`+field+`":0}`)
			assert.Equal(t, http.StatusOK, status, "%s %s=0 must be accepted", path, field)
		}
	}
}

func TestWebhookEventValidation(t *testing.T) {
	server := testsuite.InitServer(t)

	post := func(path, body string) (int, string) {
		request := testsuite.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
		request.Header.Set("Content-Type", "application/json")
		response := server.TestRequest(request)
		defer response.Body.Close() // nolint:errcheck
		b, _ := io.ReadAll(response.Body)
		return response.StatusCode, string(b)
	}
	global := func(event string) string {
		return `{"event":"` + event + `","url":"http://127.0.0.1:1/x"}`
	}
	direct := func(event string) string {
		return `{"name":"t","command":"-version","webhooks":[{"event":"` + event + `","url":"http://127.0.0.1:1/x"}]}`
	}

	cases := []struct {
		path     string
		body     func(string) string
		accepted []string
		rejected []string
	}{
		{"/api/v1/webhooks", global,
			[]string{"task.updated", "batch.finished", "preset.created", "webhook.deleted", "watchfolder.updated"},
			[]string{"task.canceled", "nope", ""}},
		{"/api/v1/tasks", direct,
			[]string{"task.created", "task.updated", "task.deleted"},
			[]string{"task.canceled", "nope", "batch.finished", "preset.created"}},
		{"/api/v1/presets", direct,
			[]string{"preset.created", "preset.updated", "preset.deleted", "task.created", "task.updated", "task.deleted"},
			[]string{"task.canceled", "bogus", "batch.finished", "webhook.created"}},
	}
	for _, c := range cases {
		for _, event := range c.accepted {
			status, resp := post(c.path, c.body(event))
			assert.Equal(t, http.StatusOK, status, "%s event %q must be accepted: %s", c.path, event, resp)
		}
		for _, event := range c.rejected {
			status, _ := post(c.path, c.body(event))
			assert.Equal(t, http.StatusUnprocessableEntity, status, "%s event %q must be rejected", c.path, event)
		}
	}
}
