package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/pkg/service"
	"github.com/welovemedia/ffmate/sev"
	"github.com/welovemedia/ffmate/sev/exceptions"
)

type TaskController struct {
	sev.Controller
	sev    *sev.Sev
	Prefix string

	taskService *service.TaskService
}

func (c *TaskController) Setup(s *sev.Sev) {
	c.taskService = &service.TaskService{
		Sev:            s,
		TaskRepository: &repository.Task{DB: s.DB()},
		PresetService: &service.PresetService{
			Sev: s,
			PresetRepository: &repository.Preset{
				DB: s.DB(),
			},
		},
		WebhookService: &service.WebhookService{
			Sev: s,
			WebhookRepository: &repository.Webhook{
				DB: s.DB(),
			},
		},
	}
	c.sev = s
	s.Gin().GET(c.Prefix+c.getEndpoint(), c.listTasks)
	s.Gin().POST(c.Prefix+c.getEndpoint(), c.addTask)
	s.Gin().GET(c.Prefix+c.getEndpoint()+"/:uuid", c.getTask)
	s.Gin().PATCH(c.Prefix+c.getEndpoint()+"/:uuid/cancel", c.cancelTask)
}

// @Summary		List all tasks
// @Description	List all existing tasks
// @Tags			tasks
// @Produce		json
// @Success		200	{object}	[]dto.Task
// @Router			/tasks [get]
func (c *TaskController) listTasks(gin *gin.Context) {
	tasks, err := c.taskService.ListTasks()
	if err != nil {
		gin.JSON(400, err)
		return
	}

	// Transform each task to its DTO
	var taskDTOs = []dto.Task{}
	for _, task := range *tasks {
		taskDTOs = append(taskDTOs, *task.ToDto())
	}

	gin.JSON(200, taskDTOs)
}

// @Summary		Add a new task
// @Description	Add a new tasks to the queue
// @Tags			tasks
// @Accept			json
// @Param			request	body	dto.NewTask	true	"new task"
// @Produce		json
// @Success		200	{object}	dto.Task
// @Router			/tasks [post]
func (c *TaskController) addTask(gin *gin.Context) {
	newTask := &dto.NewTask{}
	c.sev.Validate().Bind(gin, newTask)

	task, err := c.taskService.NewTask(newTask)
	if err != nil {
		gin.JSON(400, err)
		return
	}

	gin.JSON(200, task.ToDto())
}

// @Summary		Get single task
// @Description	Get a single task by its uuid
// @Tags			tasks
// @Param			uuid	path	string	true	"the tasks uuid"
// @Produce		json
// @Success		200	{object}	dto.Task
// @Router			/tasks/{uuid} [get]
func (c *TaskController) getTask(gin *gin.Context) {
	uuid := gin.Param("uuid")
	task, err := c.taskService.GetTaskById(uuid)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err))
		return
	}

	gin.JSON(200, task.ToDto())
}

// @Summary		Cancel a task
// @Description	Cancel a task by its uuid
// @Tags			tasks
// @Param			uuid	path	string	true	"the tasks uuid"
// @Produce		json
// @Success		200	{object}	dto.Task
// @Router			/tasks/{uuid}/cancel [patch]
func (c *TaskController) cancelTask(gin *gin.Context) {
	uuid := gin.Param("uuid")
	task, err := c.taskService.CancelTask(uuid)
	if err != nil {
		gin.JSON(400, exceptions.HttpBadRequest(err))
		return
	}

	gin.JSON(200, task.ToDto())
}

func (c *TaskController) GetName() string {
	return "task"
}

func (c *TaskController) getEndpoint() string {
	return "/v1/tasks"
}
