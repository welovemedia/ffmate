package controller

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/sev"
)

type ClientController struct {
	sev.Controller
	sev *sev.Sev

	Prefix string
}

func (c *ClientController) Setup(s *sev.Sev) {
	c.sev = s
	s.Gin().GET(c.Prefix+c.getEndpoint(), c.getClient)
}

// @Summary Get Client info
// @Description Get Client info
// @Tags client
// @Produce json
// @Success 200 {object} dto.Client
// @Router /client [get]
func (c *ClientController) getClient(gin *gin.Context) {
	config.Config().Mutex.RLock()
	defer config.Config().Mutex.RUnlock()
	gin.JSON(200, &dto.Client{
		Version: config.Config().AppVersion,
		FFmpeg:  config.Config().FFMpeg,
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	})
}

func (c *ClientController) GetName() string {
	return "client"
}

func (c *ClientController) getEndpoint() string {
	return "/v1/client"
}
