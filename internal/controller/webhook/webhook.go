package webhook

import (
	"fmt"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/exception"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/validate"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/typeutil"
)

type Service interface {
	List(page int, perPage int) (*[]model.Webhook, int64, error)
	ListExecutions(page int, perPage int) (*[]model.WebhookExecution, int64, error)
	Add(*dto.NewWebhook) (*model.Webhook, error)
	Delete(string) error
	Get(string) (*model.Webhook, error)
	Update(string, *dto.NewWebhook) (*model.Webhook, error)
}

type Controller struct {
	goyave.Component
	webhookService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	c.webhookService = server.Service(service.Webhook).(Service)
	debug.Controller.Debug("registered webhook controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Delete("/webhooks/{uuid}", c.delete)
	router.Post("/webhooks", c.add).ValidateBody(c.NewWebhookRequest)
	router.Put("/webhooks/{uuid}", c.update).ValidateBody(c.NewWebhookRequest)
	router.Get("/webhooks", c.list).ValidateQuery(validate.PaginationRequest)
	router.Get("/webhooks/executions", c.listExecutions).ValidateQuery(validate.PaginationRequest)
	router.Get("/webhooks/{uuid}", c.get)

}

// @Summary Delete a webhook
// @Description Delete a webhook by its uuid
// @Tags webhooks
// @Param uuid path string true "the webhooks uuid"
// @Produce json
// @Success 204
// @Router /webhooks/{uuid} [delete]
func (c *Controller) delete(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]
	err := c.webhookService.Delete(uuid)

	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#deleting-a-webhook"))
		return
	}

	response.Status(204)
}

// @Summary List all webhooks
// @Description List all existing webhooks
// @Tags webhooks
// @Produce json
// @Success 200 {object} []dto.Webhook
// @Router /webhooks [get]
func (c *Controller) list(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)

	webhooks, total, err := c.webhookService.List(query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#listing-all-webhooks"))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	// Transform each webhook to its DTO
	var webhookDTOs = []dto.Webhook{}
	for _, webhook := range *webhooks {
		webhookDTOs = append(webhookDTOs, *webhook.ToDto())
	}

	response.JSON(200, webhookDTOs)
}

// @Summary List all webhooks executions
// @Description List all existing webhook executions
// @Tags webhooks
// @Produce json
// @Success 200 {object} []dto.WebhookExecution
// @Router /webhooks/executions [get]
func (c *Controller) listExecutions(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)
	webhookExecutions, total, err := c.webhookService.ListExecutions(query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, ""))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	// Transform each webhook to its DTO
	var webhooksExecutionDTOs = []dto.WebhookExecution{}
	for _, webhookExecution := range *webhookExecutions {
		webhooksExecutionDTOs = append(webhooksExecutionDTOs, *webhookExecution.ToDto())
	}

	response.JSON(200, webhooksExecutionDTOs)
}

// @Summary Add a new webhook
// @Description Add a new webhook for an event
// @Tags webhooks
// @Accept json
// @Param request body dto.NewWebhook true "new webhook"
// @Produce json
// @Success 200 {object} dto.Webhook
// @Router /webhooks [post]
func (c *Controller) add(response *goyave.Response, request *goyave.Request) {
	newWebhook := typeutil.MustConvert[*dto.NewWebhook](request.Data)

	webhook, err := c.webhookService.Add(newWebhook)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#creating-a-webhook"))
		return
	}

	response.JSON(200, webhook.ToDto())
}

// @Summary Get single webhook
// @Description Get a single webhook by its uuid
// @Tags webhooks
// @Param uuid path string true "the webhooks uuid"
// @Produce json
// @Success 200 {object} dto.Webhook
// @Router /webhooks/{uuid} [get]
func (c *Controller) get(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]

	webhook, err := c.webhookService.Get(uuid)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#getting-a-single-webhook"))
		return
	}

	response.JSON(200, webhook.ToDto())
}

// @Summary Update a webhook
// @Description Update a webhook for an event
// @Tags webhooks
// @Accept json
// @Param request body dto.NewWebhook true "updated webhook"
// @Produce json
// @Success 200 {object} dto.Webhook
// @Router /webhooks/{uuid} [put]
func (c *Controller) update(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]
	newWebhook := typeutil.MustConvert[*dto.NewWebhook](request.Data)

	webhook, err := c.webhookService.Update(uuid, newWebhook)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/webhooks#updating-a-webhook"))
		return
	}

	response.JSON(200, webhook.ToDto())
}
