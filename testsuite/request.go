package testsuite

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/welovemedia/ffmate/v2/internal/cfg"
)

func NewRequest(method, target string, body io.Reader) *http.Request {
	r := httptest.NewRequest(method, target, body)
	if cfg.GetBool("ffmate.isAuth") {
		r.Header.Add("Authorization", "Basic dXNlcjpwYXNz")
	}
	return r
}
