package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/sev"
)

type WebController struct {
	sev.Controller
	sev *sev.Sev

	Prefix string
}

func (c *WebController) Setup(s *sev.Sev) {
	c.sev = s
	s.Gin().GET(c.getEndpoint(), c.serve)
}

func (c *WebController) serve(gin *gin.Context) {
	gin.AbortWithStatus(501)
}

func (c *WebController) GetName() string {
	return "web"
}

func (c *WebController) getEndpoint() string {
	return "/web"
}
