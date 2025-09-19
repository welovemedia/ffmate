package preset

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
	List(page int, perPage int) (*[]model.Preset, int64, error)
	Add(*dto.NewPreset) (*model.Preset, error)
	Delete(string) error
	Get(string) (*model.Preset, error)
	Update(string, *dto.NewPreset) (*model.Preset, error)
}

type Controller struct {
	goyave.Component
	PresetService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.PresetService = server.Service(service.Preset).(Service)
	c.Component.Init(server)
	debug.Controller.Debug("registered preset controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Delete("/presets/{uuid}", c.delete)
	router.Post("/presets", c.add).ValidateBody(c.NewPresetRequest)
	router.Put("/presets/{uuid}", c.update).ValidateBody(c.NewPresetRequest)
	router.Get("/presets", c.list).ValidateQuery(validate.PaginationRequest)
	router.Get("/presets/{uuid}", c.get)
}

// @Summary Delete a preset
// @Description Delete a preset by its uuid
// @Tags presets
// @Param uuid path string true "the presets uuid"
// @Produce json
// @Success 204
// @Router /presets/{uuid} [delete]
func (c *Controller) delete(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]
	err := c.PresetService.Delete(uuid)

	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/presets#deleting-a-preset"))
		return
	}

	response.Status(204)
}

// @Summary List all presets
// @Description	List all existing presets
// @Tags presets
// @Produce json
// @Success 200 {object} []dto.Preset
// @Router /presets [get]
func (c *Controller) list(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)

	presets, total, err := c.PresetService.List(query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/presets#listing-presets"))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	// Transform each preset to its DTO
	var presetDTOs = []dto.Preset{}
	for _, preset := range *presets {
		presetDTOs = append(presetDTOs, *preset.ToDto())
	}

	response.JSON(200, presetDTOs)
}

// @Summary Add a new preset
// @Description	Add a new preset
// @Tags presets
// @Accept json
// @Param request body dto.NewPreset true "new preset"
// @Produce json
// @Success 200 {object} dto.Preset
// @Router /presets [post]
func (c *Controller) add(response *goyave.Response, request *goyave.Request) {
	newPreset := typeutil.MustConvert[*dto.NewPreset](request.Data)

	preset, err := c.PresetService.Add(newPreset)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/presets#creating-a-preset"))
		return
	}

	response.JSON(200, preset.ToDto())
}

// @Summary Get a preset
// @Description	Get a preset
// @Tags presets
// @Produce json
// @Success 200 {object} dto.Preset
// @Router /presets/{uuid} [get]
func (c *Controller) get(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]

	preset, err := c.PresetService.Get(uuid)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/presets#getting-a-single-preset"))
		return
	}

	response.JSON(200, preset.ToDto())
}

// @Summary Update a preset
// @Description	Update a preset
// @Tags presets
// @Accept json
// @Param request body dto.NewPreset true "new preset"
// @Produce json
// @Success 200 {object} dto.Preset
// @Router /presets/{uuid} [put]
func (c *Controller) update(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]
	newPreset := typeutil.MustConvert[*dto.NewPreset](request.Data)

	preset, err := c.PresetService.Update(uuid, newPreset)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/presets#updating-a-preset"))
		return
	}

	response.JSON(200, preset.ToDto())
}
