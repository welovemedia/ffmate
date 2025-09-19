package watchfolder

import (
	"fmt"

	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/exception"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/validate"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/typeutil"
)

type Service interface {
	List(page int, perPage int) (*[]model.Watchfolder, int64, error)
	Add(*dto.NewWatchfolder) (*model.Watchfolder, error)
	Delete(string) error
	Get(string) (*model.Watchfolder, error)
	Update(string, *dto.NewWatchfolder) (*model.Watchfolder, error)
}

type Controller struct {
	goyave.Component
	watchfolderService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	c.watchfolderService = server.Service(service.Watchfolder).(Service)
	debug.Controller.Debug("registered watchfolder controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Delete("/watchfolders/{uuid}", c.delete) // deprecated (typo)
	router.Delete("/watchfolder/{uuid}", c.delete)
	router.Post("/watchfolders", c.add).ValidateBody(c.NewWatchfolderRequest) // deprecated (typo)
	router.Post("/watchfolder", c.add).ValidateBody(c.NewWatchfolderRequest)
	router.Put("/watchfolders/{uuid}", c.update).ValidateBody(c.NewWatchfolderRequest) // deprecated (typo)
	router.Put("/watchfolder/{uuid}", c.update).ValidateBody(c.NewWatchfolderRequest)
	router.Get("/watchfolders", c.list).ValidateQuery(validate.PaginationRequest) // deprecated (typo)
	router.Get("/watchfolder", c.list).ValidateQuery(validate.PaginationRequest)
	router.Get("/watchfolders/{uuid}", c.get) // deprecated (typo)
	router.Get("/watchfolder/{uuid}", c.get)
}

// @Summary Delete a watchfolder
// @Description Delete a watchfolder by its uuid
// @Tags watchfolder
// @Param uuid path string true "the watchfolders uuid"
// @Produce json
// @Success 204
// @Router /watchfolder/{uuid} [delete]
func (c *Controller) delete(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]
	err := c.watchfolderService.Delete(uuid)

	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#deleting-a-watchfolder"))
		return
	}

	response.Status(204)
}

// @Summary List all watchfolders
// @Description List all existing watchfolders
// @Tags watchfolder
// @Produce json
// @Success 200 {object} []dto.Watchfolder
// @Router /watchfolder [get]
func (c *Controller) list(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)

	watchfolder, total, err := c.watchfolderService.List(query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#listing-watchfolders"))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	// Transform each watchfolder to its DTO
	var watchfolderDTOs = []dto.Watchfolder{}
	for _, watchfolder := range *watchfolder {
		watchfolderDTOs = append(watchfolderDTOs, *watchfolder.ToDto())
	}

	response.JSON(200, watchfolderDTOs)
}

// @Summary Add a new watchfolder
// @Description Add a new watchfolder
// @Tags watchfolder
// @Accept json
// @Param request body dto.NewWatchfolder true "new watchfolder"
// @Produce json
// @Success 200 {object} dto.Watchfolder
// @Router /watchfolder [post]
func (c *Controller) add(response *goyave.Response, request *goyave.Request) {
	newWatchfolder := typeutil.MustConvert[*dto.NewWatchfolder](request.Data)

	watchfolder, err := c.watchfolderService.Add(newWatchfolder)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#creating-a-watchfolder"))
		return
	}

	response.JSON(200, watchfolder.ToDto())
}

// @Summary Get single watchfolder
// @Description	Get a single watchfolder by its uuid
// @Tags watchfolder
// @Param uuid path string true "the watchfolders uuid"
// @Produce json
// @Success 200 {object} dto.Watchfolder
// @Router /watchfolder/{uuid} [get]
func (c *Controller) get(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]

	watchfolder, err := c.watchfolderService.Get(uuid)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#getting-a-single-watchfolder"))
		return
	}

	response.JSON(200, watchfolder.ToDto())
}

// @Summary Update a watchfolder
// @Description Update a watchfolder
// @Tags watchfolder
// @Accept json
// @Param request body dto.NewWatchfolder true "new watchfolder"
// @Produce json
// @Success 200 {object} dto.Watchfolder
// @Router /watchfolder [put]
func (c *Controller) update(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]
	newWebhook := typeutil.MustConvert[*dto.NewWatchfolder](request.Data)

	watchfolder, err := c.watchfolderService.Update(uuid, newWebhook)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#updating-a-watchfolder"))
		return
	}

	response.JSON(200, watchfolder.ToDto())
}
