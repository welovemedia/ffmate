package controller

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/sev"
	"github.com/welovemedia/ffmate/sev/exceptions"
)

type AIController struct {
	sev.Controller
	sev *sev.Sev

	Prefix string
}

func (c *AIController) Setup(s *sev.Sev) {
	c.sev = s
	s.Gin().GET(c.Prefix+c.getEndpoint(), c.getAI)
}

// @Summary Get AI configuration
// @Description Get AI configuration
// @Tags ai
// @Produce json
// @Success 200 {object} dto.AI
// @Router /ai [get]
func (c *AIController) getAI(gin *gin.Context) {
	ai := config.Config().AI
	if ai == "" {
		gin.AbortWithStatusJSON(404, exceptions.HttpNotFound(errors.New("ai not configured"), ""))
		return
	}

	s := strings.Split(ai, ":")
	if len(s) != 3 {
		gin.AbortWithStatusJSON(400, exceptions.HttpBadRequest(errors.New("ai key is invalid, expected format: vendor:model:key"), ""))
		return
	}

	gin.JSON(200, &dto.AI{
		Vendor: s[0],
		Model:  s[1],
		Key:    s[2],
	})
}

func (c *AIController) GetName() string {
	return "ai"
}

func (c *AIController) getEndpoint() string {
	return "/v1/ai"
}
