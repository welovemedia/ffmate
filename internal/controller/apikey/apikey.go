package apikey

import (
	"fmt"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/exception"
	"github.com/welovemedia/ffmate/v2/internal/middleware"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/validate"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/typeutil"
)

type Service interface {
	List(page int, perPage int) (*[]model.APIKey, int64, error)
	Add(newAPIKey *dto.NewAPIKey) (*model.APIKey, string, error)
	Get(uuid string) (*model.APIKey, error)
	Rotate(uuid string) (*model.APIKey, string, error)
	Delete(uuid string) error
}

type Controller struct {
	goyave.Component
	apiKeyService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	c.apiKeyService = server.Service(service.APIKey).(Service)
	debug.Controller.Debug("registered apikey controller")
}

// RegisterRoutes registers the api key routes. If basic auth is disabled, all routes respond with 412.
// All routes are restricted to basic auth, api keys cannot manage api keys.
func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router = router.Subrouter("/apikeys")
	router.Middleware(&middleware.RequireBasicAuthMiddleware{})

	router.Get("", c.list).ValidateQuery(validate.PaginationRequest).SetMeta(middleware.MetaAccess, middleware.AccessBasicAuth)
	router.Post("", c.add).ValidateBody(c.NewAPIKeyRequest).SetMeta(middleware.MetaAccess, middleware.AccessBasicAuth)
	router.Get("/{uuid}", c.get).SetMeta(middleware.MetaAccess, middleware.AccessBasicAuth)
	router.Post("/{uuid}/rotate", c.rotate).SetMeta(middleware.MetaAccess, middleware.AccessBasicAuth)
	router.Delete("/{uuid}", c.delete).SetMeta(middleware.MetaAccess, middleware.AccessBasicAuth)
}

// @Summary List all api keys
// @Description List all existing api keys (without the secret key)
// @Tags apikeys
// @Produce json
// @Success 200 {object} []dto.APIKey
// @Failure 412 {object} exception.HTTPError
// @Router /apikeys [get]
func (c *Controller) list(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)

	apiKeys, total, err := c.apiKeyService.List(query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HTTPBadRequest(err, ""))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	var apiKeyDTOs = []dto.APIKey{}
	for _, apiKey := range *apiKeys {
		apiKeyDTOs = append(apiKeyDTOs, *apiKey.ToDTO())
	}

	response.JSON(200, apiKeyDTOs)
}

// @Summary Add a new api key
// @Description Add a new api key. The secret key is only returned once in this response.
// @Tags apikeys
// @Accept json
// @Param request body dto.NewAPIKey true "new api key"
// @Produce json
// @Success 200 {object} dto.CreatedAPIKey
// @Failure 412 {object} exception.HTTPError
// @Router /apikeys [post]
func (c *Controller) add(response *goyave.Response, request *goyave.Request) {
	newAPIKey := typeutil.MustConvert[*dto.NewAPIKey](request.Data)

	apiKey, key, err := c.apiKeyService.Add(newAPIKey)
	if err != nil {
		response.JSON(400, exception.HTTPBadRequest(err, ""))
		return
	}

	response.JSON(200, &dto.CreatedAPIKey{APIKey: *apiKey.ToDTO(), Key: key})
}

// @Summary Get single api key
// @Description Get a single api key by its uuid (without the secret key)
// @Tags apikeys
// @Param uuid path string true "the api keys uuid"
// @Produce json
// @Success 200 {object} dto.APIKey
// @Failure 412 {object} exception.HTTPError
// @Router /apikeys/{uuid} [get]
func (c *Controller) get(response *goyave.Response, request *goyave.Request) {
	apiKey, err := c.apiKeyService.Get(request.RouteParams["uuid"])
	if err != nil {
		response.JSON(400, exception.HTTPBadRequest(err, ""))
		return
	}

	response.JSON(200, apiKey.ToDTO())
}

// @Summary Rotate an api key
// @Description Replace the secret key of an existing api key. Name, access and uuid are kept, the previous key is invalid immediately. The new secret key is only returned once in this response.
// @Tags apikeys
// @Param uuid path string true "the api keys uuid"
// @Produce json
// @Success 200 {object} dto.CreatedAPIKey
// @Failure 412 {object} exception.HTTPError
// @Router /apikeys/{uuid}/rotate [post]
func (c *Controller) rotate(response *goyave.Response, request *goyave.Request) {
	apiKey, key, err := c.apiKeyService.Rotate(request.RouteParams["uuid"])
	if err != nil {
		response.JSON(400, exception.HTTPBadRequest(err, ""))
		return
	}

	response.JSON(200, &dto.CreatedAPIKey{APIKey: *apiKey.ToDTO(), Key: key})
}

// @Summary Delete an api key
// @Description Delete (revoke) an api key by its uuid
// @Tags apikeys
// @Param uuid path string true "the api keys uuid"
// @Produce json
// @Success 204
// @Failure 412 {object} exception.HTTPError
// @Router /apikeys/{uuid} [delete]
func (c *Controller) delete(response *goyave.Response, request *goyave.Request) {
	if err := c.apiKeyService.Delete(request.RouteParams["uuid"]); err != nil {
		response.JSON(400, exception.HTTPBadRequest(err, ""))
		return
	}

	response.Status(204)
}
