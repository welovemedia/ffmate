package internal

import (
	"embed"

	"github.com/gin-contrib/cors"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/welovemedia/ffmate/internal/controller"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/metrics"
	"github.com/welovemedia/ffmate/internal/middleware"
	"github.com/welovemedia/ffmate/internal/queue"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/sev"
)

var prefix = "/api"

func Init(s *sev.Sev, concurrentTasks uint, frontend embed.FS) {
	// setup cors
	s.Gin().Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// setup repositories
	(&repository.Task{DB: s.DB()}).Setup()
	(&repository.Webhook{DB: s.DB()}).Setup()
	(&repository.Preset{DB: s.DB()}).Setup()

	// setup metrics
	metrics := &metrics.Metrics{}
	for name, gauge := range metrics.Gauges() {
		s.Metrics().RegisterGauge(name, gauge)
	}

	// setup middlewares
	s.RegisterMiddleware("404", middleware.E404)
	s.RegisterMiddleware("debugo", middleware.Debugo)
	s.RegisterMiddleware("version", middleware.Version)

	// setup controllers
	s.RegisterController(&controller.TaskController{Prefix: prefix})
	s.RegisterController(&controller.WebhookController{Prefix: prefix})
	s.RegisterController(&controller.PresetController{Prefix: prefix})
	if !config.Config().Headless {
		s.RegisterController(&controller.WebController{Prefix: prefix, Frontend: frontend})
	}
	s.RegisterController(&controller.DebugController{Prefix: prefix})
	s.RegisterController(&controller.VersionController{Prefix: prefix})
	s.RegisterController(&controller.WebsocketController{Prefix: prefix})

	// Initialize queue processor
	(&queue.Queue{
		Sev:            s,
		TaskRepository: &repository.Task{DB: s.DB()},
		TaskService: &service.TaskService{
			Sev: s,
			TaskRepository: &repository.Task{
				DB: s.DB(),
			},
			WebhookService: &service.WebhookService{
				Sev: s,
				WebhookRepository: &repository.Webhook{
					DB: s.DB(),
				},
			},
			PresetService: &service.PresetService{
				Sev: s,
				PresetRepository: &repository.Preset{
					DB: s.DB(),
				},
			},
			WebsocketService: &service.WebsocketService{},
		},
		WebhookService: &service.WebhookService{
			Sev: s,
			WebhookRepository: &repository.Webhook{
				DB: s.DB(),
			},
		},
		WebsocketService:   &service.WebsocketService{},
		MaxConcurrentTasks: concurrentTasks}).Init()
}
