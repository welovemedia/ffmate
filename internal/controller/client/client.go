package client

import (
	"fmt"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/exception"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/typeutil"
)

type Service interface {
	List(int, int) (*[]model.Client, int64, error)
}

type Controller struct {
	goyave.Component
	clientService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	c.clientService = c.Server().Service(service.Client).(Service)
	debug.Controller.Debug("registered client controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Get("/clients", c.list)
}

// @Summary List all clients
// @Description List all existing client
// @Tags clients
// @Param page query int false "the page of a pagination request (min 0)"
// @Param perPage query int false "the amount of results of a pagination request (min 1; max: 100)"
// @Produce json
// @Success 200 {object} []dto.Client
// @Router /clients [get]
func (c *Controller) list(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)

	clients, total, err := c.clientService.List(query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/settings#load-settings"))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	// Transform each client to its DTO
	var clientDTOs = []dto.Client{}
	for _, client := range *clients {
		clientDTOs = append(clientDTOs, *client.ToDto())
	}

	response.JSON(200, clientDTOs)
}
