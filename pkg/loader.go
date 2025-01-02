package pkg

import (
	"github.com/welovemedia/ffmate/pkg/controller"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/metrics"
	"github.com/welovemedia/ffmate/pkg/middleware"
	"github.com/welovemedia/ffmate/pkg/queue"
	"github.com/welovemedia/ffmate/sev"
)

var prefix = "/api"

func Init(s *sev.Sev, concurrentTasks uint) {
	// setup repositories
	(&repository.Task{DB: s.DB()}).Setup()
	(&repository.Webhook{DB: s.DB()}).Setup()

	// setup metrics
	metrics := &metrics.Metrics{}
	for name, gauge := range metrics.Gauges() {
		s.Metrics().RegisterGauge(name, gauge)
	}

	// setup middlewares
	s.RegisterMiddleware("404", middleware.E404)
	s.RegisterMiddleware("version", middleware.Version)

	// setup controllers
	s.RegisterController(&controller.TaskController{Prefix: prefix})
	s.RegisterController(&controller.WebhookController{Prefix: prefix})
	s.RegisterController(&controller.WebController{Prefix: prefix})
	s.RegisterController(&controller.VersionController{Prefix: prefix})

	// Initialize queue processor
	(&queue.Queue{Sev: s, TaskRepository: &repository.Task{DB: s.DB()}, MaxConcurrentTasks: concurrentTasks}).Init()
}
