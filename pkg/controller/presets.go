package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/pkg/service"
	"github.com/welovemedia/ffmate/sev"
	"github.com/welovemedia/ffmate/sev/exceptions"
)

type PresetController struct {
	sev.Controller
	sev           *sev.Sev
	presetService *service.PresetService

	Prefix string
}

func (c *PresetController) Setup(s *sev.Sev) {
	c.presetService = &service.PresetService{
		Sev:              s,
		PresetRepository: &repository.Preset{DB: s.DB()},
		WebhookService: &service.WebhookService{
			Sev:               s,
			WebhookRepository: &repository.Webhook{DB: s.DB()},
		},
	}
	c.sev = s
	s.Gin().DELETE(c.Prefix+c.getEndpoint()+"/:uuid", c.deletePreset)
	s.Gin().POST(c.Prefix+c.getEndpoint(), c.addPreset)
	s.Gin().GET(c.Prefix+c.getEndpoint(), c.listPresets)
}

// @Summary		Delete a preset
// @Description	Delete a preset by its uuid
// @Tags			presets
// @Param			uuid	path	string	true	"the presets uuid"
// @Produce		json
// @Success		204
// @Router			/presets/{uuid} [delete]
func (c *PresetController) deletePreset(gin *gin.Context) {
	uuid := gin.Param("uuid")
	err := c.presetService.DeletePreset(uuid)

	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err))
		return
	}

	gin.AbortWithStatus(204)
}

// @Summary		List all presets
// @Description	List all existing presets
// @Tags			presets
// @Produce		json
// @Success		200	{object}	[]dto.Preset
// @Router			/presets [get]
func (c *PresetController) listPresets(gin *gin.Context) {
	presets, err := c.presetService.ListPresets()
	if err != nil {
		gin.JSON(400, err)
		return
	}

	// Transform each preset to its DTO
	var presetDTOs = []dto.Preset{}
	for _, preset := range *presets {
		presetDTOs = append(presetDTOs, *preset.ToDto())
	}

	gin.JSON(200, presetDTOs)
}

// @Summary		Add a new preset
// @Description	Add a new preset
// @Tags			presets
// @Accept			json
// @Param			request	body	dto.NewPreset	true	"new preset"
// @Produce		json
// @Success		200	{object}	dto.Preset
// @Router			/presets [post]
func (c *PresetController) addPreset(gin *gin.Context) {
	newPreset := &dto.NewPreset{}
	c.sev.Validate().Bind(gin, newPreset)

	preset, err := c.presetService.NewPreset(newPreset)
	if err != nil {
		gin.JSON(400, err)
		return
	}

	gin.JSON(200, preset.ToDto())
}

func (c *PresetController) GetName() string {
	return "preset"
}

func (c *PresetController) getEndpoint() string {
	return "/v1/presets"
}
