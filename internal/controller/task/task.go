package task

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
	List(page int, perPage int) (*[]model.Task, int64, error)
	GetBatch(uuid string, page int, perPage int) (*dto.Batch, int64, error)
	Add(task *dto.NewTask, source dto.TaskSource, batch string) (*model.Task, error)
	AddBatch(btach *dto.NewBatch, source dto.TaskSource) (*dto.Batch, error)
	Delete(uuid string) error
	Get(uuid string) (*model.Task, error)
	Cancel(uuid string) (*model.Task, error)
	Restart(uuid string) (*model.Task, error)
}

type Controller struct {
	goyave.Component
	taskService Service
}

func (c *Controller) Init(server *goyave.Server) {
	c.Component.Init(server)
	c.taskService = server.Service(service.Task).(Service)
	debug.Controller.Debug("registered task controller")
}

func (c *Controller) RegisterRoutes(router *goyave.Router) {
	router.Delete("/tasks/{uuid}", c.delete)
	router.Post("/tasks", c.add).ValidateBody(c.NewTaskRequest)
	router.Get("/tasks", c.list).ValidateQuery(validate.PaginationRequest)
	router.Get("/tasks/{uuid}", c.get)
	router.Patch("/tasks/{uuid}/cancel", c.cancel)
	router.Patch("/tasks/{uuid}/restart", c.restart)

	router.Post("/batches", c.addBatch)
	router.Get("/batches/{uuid}", c.getBatch).ValidateQuery(validate.PaginationRequest)
}

// @Summary Delete a task
// @Description Delete a task by its uuid
// @Tags tasks
// @Param uuid path string true "the tasks uuid"
// @Produce json
// @Success 204
// @Router /tasks/{uuid} [delete]
func (c *Controller) delete(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]
	err := c.taskService.Delete(uuid)

	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#deleting-a-task"))
		return
	}

	response.Status(204)
}

// @Summary List all tasks
// @Description List all existing tasks
// @Tags tasks
// @Param page query int false "the page of a pagination request (min 0)"
// @Param perPage query int false "the amount of results of a pagination request (min 1; max: 100)"
// @Produce json
// @Success 200 {object} []dto.Task
// @Router /tasks [get]
func (c *Controller) list(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)

	tasks, total, err := c.taskService.List(query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#monitoring-a-task"))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	// Transform each preset to its DTO
	var taskDTOs = []dto.Task{}
	for _, task := range *tasks {
		taskDTOs = append(taskDTOs, *task.ToDto())
	}

	response.JSON(200, taskDTOs)
}

// @Summary Get a btach
// @Description	Get a batch by uuid
// @Tags tasks
// @Param uuid path string true "the batch uuid"
// @Produce json
// @Success 200 {object} dto.Batch
// @Router /batches/{uuid} [get]
func (c *Controller) getBatch(response *goyave.Response, request *goyave.Request) {
	query := typeutil.MustConvert[*dto.Pagination](request.Query)
	uuid := request.RouteParams["uuid"]

	batch, total, err := c.taskService.GetBatch(uuid, query.Page.Default(0), query.PerPage.Default(100))
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#monitoring-all-tasks"))
		return
	}

	response.Header().Set("X-Total", fmt.Sprintf("%d", total))

	response.JSON(200, batch)
}

// @Summary Add a new task
// @Description	Add a new tasks to the queue
// @Tags tasks
// @Accept json
// @Param request body dto.NewTask true "new task"
// @Produce json
// @Success 200 {object} dto.Task
// @Router /tasks [post]
func (c *Controller) add(response *goyave.Response, request *goyave.Request) {
	newTask := typeutil.MustConvert[*dto.NewTask](request.Data)

	preset, err := c.taskService.Add(newTask, dto.API, "")
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#creating-a-task"))
		return
	}

	response.JSON(200, preset.ToDto())
}

// @Summary Add a batch of tasks
// @Description	Add a batch of new tasks to the queue
// @Tags tasks
// @Accept json
// @Param request body []dto.NewBatch true "new batch"
// @Produce json
// @Success 200 {object} []dto.Batch
// @Router /batches [post]
func (c *Controller) addBatch(response *goyave.Response, request *goyave.Request) {
	newBatch := typeutil.MustConvert[*dto.NewBatch](request.Data)

	batch, err := c.taskService.AddBatch(newBatch, dto.API)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#submitting-multiple-tasks-as-a-batch"))
		return
	}

	response.JSON(200, batch)
}

// @Summary Get single task
// @Description	Get a single task by its uuid
// @Tags tasks
// @Param uuid path string true "the tasks uuid"
// @Produce json
// @Success 200 {object} dto.Task
// @Router /tasks/{uuid} [get]
func (c *Controller) get(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]

	task, err := c.taskService.Get(uuid)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#monitoring-a-task"))
		return
	}

	response.JSON(200, task.ToDto())
}

// @Summary Cancel a task
// @Description Cancel a task by its uuid
// @Tags tasks
// @Param uuid path string true "the tasks uuid"
// @Produce json
// @Success 200 {object} dto.Task
// @Router /tasks/{uuid}/cancel [patch]
func (c *Controller) cancel(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]

	task, err := c.taskService.Cancel(uuid)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#canceling-a-task"))
		return
	}

	response.JSON(200, task.ToDto())
}

// @Summary Restart a task
// @Description Restart a task by its uuid
// @Tags tasks
// @Param uuid path string true "the tasks uuid"
// @Produce json
// @Success 200 {object} dto.Task
// @Router /tasks/{uuid}/restart [patch]
func (c *Controller) restart(response *goyave.Response, request *goyave.Request) {
	uuid := request.RouteParams["uuid"]

	task, err := c.taskService.Restart(uuid)
	if err != nil {
		response.JSON(400, exception.HttpBadRequest(err, "https://docs.ffmate.io/docs/tasks#restarting-a-task"))
		return
	}

	response.JSON(200, task.ToDto())
}
