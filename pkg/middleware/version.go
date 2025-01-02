package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/sev"
)

func Version(c *gin.Context, s *sev.Sev) {
	c.Writer.Header().Set("X-Server", s.AppName()+"/v"+s.AppVersion())
	c.Next()
}
