package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/sev"
	"github.com/yosev/debugo"
)

var debug = debugo.New("gin")

func Debugo(c *gin.Context, s *sev.Sev) {
	debug.Debugf("%s \"%s\"", c.Request.Method, c.Request.URL)
	c.Next()
}
