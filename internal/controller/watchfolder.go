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

type WatchfolderController struct {
	sev.Controller
	sev *sev.Sev

	Prefix string
}

func (c *WatchfolderController) Setup(s *sev.Sev) {
	c.sev = s
	s.Gin().DELETE(c.Prefix+c.getEndpoint()+"/:uuid", c.deleteWatchfolder)
	s.Gin().PUT(c.Prefix+c.getEndpoint()+"/:uuid", c.updateWatchfolder)
	s.Gin().POST(c.Prefix+c.getEndpoint(), c.addWatchfolder)
	s.Gin().GET(c.Prefix+c.getEndpoint(), interceptor.PageLimit, c.listWatchfolders)
	s.Gin().GET(c.Prefix+c.getEndpoint()+"/:uuid", c.getWatchfolder)
}

// @Summary Get single watchfolder
// @Description	Get a single watchfolder by its uuid
// @Tags watchfolders
// @Param uuid path string true "the watchfolders uuid"
// @Produce json
// @Success 200 {object} dto.Watchfolder
// @Router /watchfolder/{uuid} [get]
func (c *WatchfolderController) getWatchfolder(gin *gin.Context) {
	uuid := gin.Param("uuid")
	watchfolder, err := service.WatchfolderService().GetWatchfolderById(uuid)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#getting-a-single-watchfolder"))
		return
	}

	gin.JSON(200, watchfolder.ToDto())
}

// @Summary Delete a watchfolder
// @Description Delete a watchfolder by its uuid
// @Tags watchfolders
// @Param uuid path string true "the watchfolders uuid"
// @Produce json
// @Success 204
// @Router /watchfolders/{uuid} [delete]
func (c *WatchfolderController) deleteWatchfolder(gin *gin.Context) {
	uuid := gin.Param("uuid")
	err := service.WatchfolderService().DeleteWatchfolder(uuid)

	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#deleting-a-watchfolder"))
		return
	}

	gin.AbortWithStatus(204)
}

// @Summary List all watchfolders
// @Description List all existing watchfolders
// @Tags watchfolders
// @Produce json
// @Success 200 {object} []dto.Watchfolder
// @Router /watchfolders [get]
func (c *WatchfolderController) listWatchfolders(gin *gin.Context) {
	watchfolders, total, err := service.WatchfolderService().ListWatchfolders(gin.GetInt("page"), gin.GetInt("perPage"))
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#listing-watchfolders"))
		return
	}

	gin.Header("X-Total", fmt.Sprintf("%d", total))

	// Transform each watchfolder to its DTO
	var watchfoldersDTOs = []dto.Watchfolder{}
	for _, watchfolder := range *watchfolders {
		watchfoldersDTOs = append(watchfoldersDTOs, *watchfolder.ToDto())
	}

	gin.JSON(200, watchfoldersDTOs)
}

// @Summary Add a new watchfolder
// @Description Add a new watchfolder
// @Tags watchfolders
// @Accept json
// @Param request body dto.NewWatchfolder true "new watchfolder"
// @Produce json
// @Success 200 {object} dto.Watchfolder
// @Router /watchfolder [post]
func (c *WatchfolderController) addWatchfolder(gin *gin.Context) {
	newWatchfolder := &dto.NewWatchfolder{}
	if !c.sev.Validate().Bind(gin, newWatchfolder) {
		return
	}

	watchfolder, err := service.WatchfolderService().NewWatchfolder(newWatchfolder)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#creating-a-watchfolder"))
		return
	}

	gin.JSON(200, watchfolder.ToDto())
}

// @Summary Update a watchfolder
// @Description Update a watchfolder
// @Tags watchfolders
// @Accept json
// @Param request body dto.NewWatchfolder true "new watchfolder"
// @Produce json
// @Success 200 {object} dto.Watchfolder
// @Router /watchfolder [put]
func (c *WatchfolderController) updateWatchfolder(gin *gin.Context) {
	uuid := gin.Param("uuid")
	newWatchfolder := &dto.NewWatchfolder{}
	if !c.sev.Validate().Bind(gin, newWatchfolder) {
		return
	}

	watchfolder, err := service.WatchfolderService().UpdateWatchfolder(uuid, newWatchfolder)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err, "https://docs.ffmate.io/docs/watchfolder#updating-a-watchfolder"))
		return
	}

	gin.JSON(200, watchfolder.ToDto())
}

func (c *WatchfolderController) GetName() string {
	return "watchfolder"
}

func (c *WatchfolderController) getEndpoint() string {
	return "/v1/watchfolders"
}
