package testsuite

import (
	"testing"

	"github.com/welovemedia/ffmate/v2/internal/controller"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/middleware"
	"github.com/welovemedia/ffmate/v2/internal/service"
	clientService "github.com/welovemedia/ffmate/v2/internal/service/client"
	"github.com/welovemedia/ffmate/v2/internal/service/ffmpeg"
	presetService "github.com/welovemedia/ffmate/v2/internal/service/preset"
	settingsSvc "github.com/welovemedia/ffmate/v2/internal/service/settings"
	taskService "github.com/welovemedia/ffmate/v2/internal/service/task"
	"github.com/welovemedia/ffmate/v2/internal/service/telemetry"
	watchfolderService "github.com/welovemedia/ffmate/v2/internal/service/watchfolder"
	webhookService "github.com/welovemedia/ffmate/v2/internal/service/webhook"
	websocketService "github.com/welovemedia/ffmate/v2/internal/service/websocket"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/testutil"
)

func InitServer(t *testing.T) *testutil.TestServer {
	server := testserver.New(t)

	// setup repositories
	presetRepository := (&repository.Preset{DB: server.DB()}).Setup()
	webhookRepository := (&repository.Webhook{DB: server.DB()}).Setup()
	webhookExecutionRepository := (&repository.WebhookExecution{DB: server.DB()}).Setup()
	watchfolderRepository := (&repository.Watchfolder{DB: server.DB()}).Setup()
	taskRepository := (&repository.Task{DB: server.DB()}).Setup()
	clientRepository := (&repository.Client{DB: server.DB()}).Setup()
	settingsRepository := (&repository.Settings{DB: server.DB()}).Setup()

	// setup and register services
	settingsSvc := settingsSvc.NewService(settingsRepository)
	telemetrySvc := telemetry.NewService(server.Config(), server.DB())
	ffmpegSvc := ffmpeg.NewService()
	websocketSvc := websocketService.NewService(server.DB())
	clientSvc := clientService.NewService(clientRepository, "test-1.0.0", websocketSvc)
	webhookSvc := webhookService.NewService(webhookRepository, webhookExecutionRepository, server.Config(), websocketSvc)
	presetSvc := presetService.NewService(presetRepository, webhookSvc, websocketSvc)
	taskSvc := taskService.NewService(taskRepository, presetSvc, webhookSvc, websocketSvc, ffmpegSvc, false)
	watchfolderSvc := watchfolderService.NewService(watchfolderRepository, webhookSvc, websocketSvc, taskSvc)
	for _, svc := range map[string]goyave.Service{
		service.FFMpeg:      ffmpegSvc,
		service.Telemetry:   telemetrySvc,
		service.Websocket:   websocketSvc,
		service.Webhook:     webhookSvc,
		service.Preset:      presetSvc,
		service.Watchfolder: watchfolderSvc,
		service.Task:        taskSvc,
		service.Settings:    settingsSvc,
		service.Client:      clientSvc,
	} {
		server.RegisterService(svc)
	}

	server.RegisterRoutes(controller.Register)
	server.RegisterRoutes(middleware.Register)

	return server
}
