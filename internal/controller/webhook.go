package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/interceptor"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/sev"
	"github.com/welovemedia/ffmate/sev/exceptions"
)

type WebhookController struct {
	sev.Controller
	sev *sev.Sev

	Prefix string
}

func (c *WebhookController) Setup(s *sev.Sev) {
	c.sev = s
	s.Gin().DELETE(c.Prefix+c.getEndpoint()+"/:uuid", c.deleteWebhook)
	s.Gin().POST(c.Prefix+c.getEndpoint(), c.addWebhook)
	s.Gin().PUT(c.Prefix+c.getEndpoint()+"/:uuid", c.updateWebhook)
	s.Gin().GET(c.Prefix+c.getEndpoint(), interceptor.PageLimit, c.listWebhooks)
	s.Gin().GET(c.Prefix+c.getEndpoint()+"/:uuid", c.getWebhook)
}

// @Summary Get single webhook
// @Description Get a single webhook by its uuid
// @Tags webhooks
// @Param uuid path string true "the webhooks uuid"
// @Produce json
// @Success 200 {object} dto.Webhook
// @Router /webhooks/{uuid} [get]
func (c *WebhookController) getWebhook(gin *gin.Context) {
	uuid := gin.Param("uuid")
	webhook, err := service.WebhookService().GetWebhookById(uuid)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#getting-a-single-webhook"))
		return
	}

	gin.JSON(200, webhook.ToDto())
}

// @Summary Delete a webhook
// @Description Delete a webhook by its uuid
// @Tags webhooks
// @Param uuid path string true "the webhooks uuid"
// @Produce json
// @Success 204
// @Router /webhooks/{uuid} [delete]
func (c *WebhookController) deleteWebhook(gin *gin.Context) {
	uuid := gin.Param("uuid")
	err := service.WebhookService().DeleteWebhook(uuid)

	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#deleting-a-webhook"))
		return
	}

	gin.AbortWithStatus(204)
}

// @Summary List all webhooks
// @Description List all existing webhooks
// @Tags webhooks
// @Produce json
// @Success 200 {object} []dto.Webhook
// @Router /webhooks [get]
func (c *WebhookController) listWebhooks(gin *gin.Context) {
	webhooks, total, err := service.WebhookService().ListWebhooks(gin.GetInt("page"), gin.GetInt("perPage"))
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#listing-all-webhooks"))
		return
	}

	gin.Header("X-Total", fmt.Sprintf("%d", total))

	// Transform each webhook to its DTO
	var webhooksDTOs = []dto.Webhook{}
	for _, webhook := range *webhooks {
		webhooksDTOs = append(webhooksDTOs, *webhook.ToDto())
	}

	gin.JSON(200, webhooksDTOs)
}

// @Summary Update a webhook
// @Description Update a webhook for an event
// @Tags webhooks
// @Accept json
// @Param request body dto.NewWebhook true "updated webhook"
// @Produce json
// @Success 200 {object} dto.Webhook
// @Router /webhooks/{uuid} [put]
func (c *WebhookController) updateWebhook(gin *gin.Context) {
	uuid := gin.Param("uuid")
	newWebhook := &dto.NewWebhook{}
	if !c.sev.Validate().Bind(gin, newWebhook) {
		return
	}

	webhook, err := service.WebhookService().UpdateWebhook(uuid, newWebhook)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#updating-a-webhook"))
		return
	}

	gin.JSON(200, webhook.ToDto())
}

// @Summary Add a new webhook
// @Description Add a new webhook for an event
// @Tags webhooks
// @Accept json
// @Param request body dto.NewWebhook true "new webhook"
// @Produce json
// @Success 200 {object} dto.Webhook
// @Router /webhooks [post]
func (c *WebhookController) addWebhook(gin *gin.Context) {
	newWebhook := &dto.NewWebhook{}
	if !c.sev.Validate().Bind(gin, newWebhook) {
		return
	}

	webhook, err := service.WebhookService().NewWebhook(newWebhook)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#creating-a-webhook"))
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
