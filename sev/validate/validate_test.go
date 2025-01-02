package validate

import (
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/assert.v1"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestQuery(t *testing.T) {
	v := &Validate{}
	fn := v.Query("url", URL)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// TODO find a way to injec a query parameter

	fn(c)

	assert.Equal(t, c.IsAborted(), true)
}
