package settings

import (
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/exception"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/typeutil"
)

type Service interface {
	Load() (*model.Settings, error)
	Store(settings *dto.Settings) (*model.Settings, error)
}

type Controller struct {
	goyave.Component
	settingsService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	c.settingsService = server.Service(service.Settings).(Service)
	debug.Controller.Debug("registered settings controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Get("/settings", c.load)
	router.Post("/settings", c.save)
}

// @Summary Get all settings
// @Description Get all existing settings
// @Tags settings
// @Produce json
// @Success 200 {object} dto.Settings
// @Router /settings [get]
func (c *Controller) load(response *goyave.Response, _ *goyave.Request) {
	settings, err := c.settingsService.Load()
	if err != nil {
		response.JSON(400, exception.HTTPBadRequest(err, "https://docs.ffmate.io/docs/settings#load-settings"))
		return
	}

	response.JSON(200, settings.ToDTO())
}

// @Summary Save all settings
// @Description	Save all existing settings
// @Tags settings
// @Accept json
// @Param request body dto.Settings true "save settings"
// @Produce json
// @Success 200 {object} dto.Settings
// @Router /settings [post]
func (c *Controller) save(response *goyave.Response, request *goyave.Request) {
	newSettings := typeutil.MustConvert[*dto.Settings](request.Data)

	settings, err := c.settingsService.Store(newSettings)
	if err != nil {
		response.JSON(400, exception.HTTPBadRequest(err, "https://docs.ffmate.io/docs/settings#save-settings"))
		return
	}

	response.JSON(200, settings.ToDTO())
}
