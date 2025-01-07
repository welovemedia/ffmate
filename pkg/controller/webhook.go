package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/pkg/service"
	"github.com/welovemedia/ffmate/sev"
	"github.com/welovemedia/ffmate/sev/exceptions"
)

type WebhookController struct {
	sev.Controller
	sev            *sev.Sev
	webhookService *service.WebhookService

	Prefix string
}

func (c *WebhookController) Setup(s *sev.Sev) {
	c.webhookService = &service.WebhookService{Sev: s, WebhookRepository: &repository.Webhook{DB: s.DB()}}
	c.sev = s
	s.Gin().DELETE(c.Prefix+c.getEndpoint()+"/:uuid", c.deleteWebhook)
	s.Gin().POST(c.Prefix+c.getEndpoint(), c.addWebhook)
	s.Gin().GET(c.Prefix+c.getEndpoint(), c.listWebhooks)
}

// @Summary		Delete a webhook
// @Description	Delete a webhook by its uuid
// @Tags			webhooks
// @Param			uuid	path	string	true	"the webhooks uuid"
// @Produce		json
// @Success		204
// @Router			/webhooks/{uuid} [delete]
func (c *WebhookController) deleteWebhook(gin *gin.Context) {
	uuid := gin.Param("uuid")
	err := c.webhookService.DeleteWebhook(uuid)

	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err))
		return
	}

	gin.AbortWithStatus(204)
}

// @Summary		List all webhooks
// @Description	List all existing webhooks
// @Tags			webhooks
// @Produce		json
// @Success		200	{object}	[]dto.Webhook
// @Router			/webhooks [get]
func (c *WebhookController) listWebhooks(gin *gin.Context) {
	webhooks, err := c.webhookService.ListWebhooks()
	if err != nil {
		gin.JSON(400, err)
		return
	}

	// Transform each webhook to its DTO
	var webhooksDTOs = []dto.Webhook{}
	for _, webhook := range *webhooks {
		webhooksDTOs = append(webhooksDTOs, *webhook.ToDto())
	}

	gin.JSON(200, webhooksDTOs)
}

// @Summary		Add a new webhook
// @Description	Add a new webhook for an event
// @Tags			webhooks
// @Accept			json
// @Param			request	body	dto.NewWebhook	true	"new webhook"
// @Produce		json
// @Success		200	{object}	dto.Webhook
// @Router			/webhooks [post]
func (c *WebhookController) addWebhook(gin *gin.Context) {
	newWebhook := &dto.NewWebhook{}
	c.sev.Validate().Bind(gin, newWebhook)

	webhook, err := c.webhookService.NewWebhook(newWebhook)
	if err != nil {
		gin.JSON(400, err)
		return
	}

	gin.JSON(200, webhook.ToDto())
}

func (c *WebhookController) GetName() string {
	return "webhook"
}

func (c *WebhookController) getEndpoint() string {
	return "/v1/webhooks"
}
