package internal

import (
	"embed"

	"github.com/gin-contrib/cors"
	"github.com/welovemedia/ffmate/internal/controller"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/metrics"
	"github.com/welovemedia/ffmate/internal/middleware"
	"github.com/welovemedia/ffmate/internal/queue"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/watchfolder"
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
	(&repository.Watchfolder{DB: s.DB()}).Setup()

	// setup metrics
	metrics := &metrics.Metrics{}
	for name, gauge := range metrics.Gauges() {
		s.Metrics().RegisterGauge(name, gauge)
	}
	for name, gauge := range metrics.GaugesVec() {
		s.Metrics().RegisterGaugeVec(name, gauge)
	}

	// setup middlewares
	s.RegisterMiddleware("404", middleware.E404)
	s.RegisterMiddleware("debugo", middleware.Debugo)
	s.RegisterMiddleware("version", middleware.Version)

	// setup services
	service.Init(s)

	// setup controllers
	s.RegisterController(&controller.TaskController{Prefix: prefix})
	s.RegisterController(&controller.WebhookController{Prefix: prefix})
	s.RegisterController(&controller.PresetController{Prefix: prefix})
	s.RegisterController(&controller.WatchfolderController{Prefix: prefix})
	s.RegisterController(&controller.WebController{Prefix: prefix, Frontend: frontend})
	s.RegisterController(&controller.DebugController{Prefix: prefix})
	s.RegisterController(&controller.VersionController{Prefix: prefix})
	s.RegisterController(&controller.WebsocketController{Prefix: prefix})
	s.RegisterController(&controller.UmamiController{Prefix: prefix})
	s.RegisterController(&controller.AIController{Prefix: prefix})

	// Initialize queue processor
	(&queue.Queue{
		Sev:                s,
		TaskRepository:     &repository.Task{DB: s.DB()},
		MaxConcurrentTasks: concurrentTasks}).Init()

	// Initialize watchfolder processor
	(&watchfolder.Watchfolder{
		Sev:                   s,
		WatchfolderRepository: &repository.Watchfolder{DB: s.DB()}}).Init()
}
