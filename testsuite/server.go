package testsuite

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/controller/client"
	"github.com/welovemedia/ffmate/v2/internal/controller/debug"
	"github.com/welovemedia/ffmate/v2/internal/controller/health"
	"github.com/welovemedia/ffmate/v2/internal/controller/preset"
	"github.com/welovemedia/ffmate/v2/internal/controller/prometheus"
	"github.com/welovemedia/ffmate/v2/internal/controller/settings"
	"github.com/welovemedia/ffmate/v2/internal/controller/swagger"
	"github.com/welovemedia/ffmate/v2/internal/controller/task"
	"github.com/welovemedia/ffmate/v2/internal/controller/ui"
	"github.com/welovemedia/ffmate/v2/internal/controller/umami"
	"github.com/welovemedia/ffmate/v2/internal/controller/version"
	"github.com/welovemedia/ffmate/v2/internal/controller/watchfolder"
	"github.com/welovemedia/ffmate/v2/internal/controller/webhook"
	"github.com/welovemedia/ffmate/v2/internal/controller/websocket"
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
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
	"goyave.dev/goyave/v5/middleware/parse"
	"goyave.dev/goyave/v5/util/testutil"
	ws "goyave.dev/goyave/v5/websocket"
)

var c = map[string]any{
	"app": map[string]any{
		"name":    "ffmate",
		"version": "test-1.0.0",
	},
	"database": map[string]any{
		"connection": "sqlite3",
		"name":       ":memory:",
	},
}

func InitServer(t *testing.T) *testutil.TestServer {
	b, _ := json.Marshal(c)
	conf, _ := config.LoadJSON(string(b))

	if !cfg.Has("ffmate.identifier") {
		cfg.Set("ffmate.identifier", "test-client")
	}
	if !cfg.Has("ffmate.labels") {
		cfg.Set("ffmate.labels", []string{"test-label-1", "test-label-2", "test-label-3"})
	}

	cfg.Set("ffmate.session", uuid.NewString())
	cfg.Set("ffmate.ffmpeg", "ffmpeg")
	cfg.Set("ffmate.maxConcurrentTasks", 3)
	cfg.Set("ffmate.isTray", false)
	cfg.Set("ffmate.isCluster", false)
	cfg.Set("ffmate.isFFmpeg", false)
	cfg.Set("ffmate.debug", "")
	cfg.Set("ffmate.isDocker", false)
	cfg.Set("ffmate.isCluster", false)

	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: conf})

	// add global parsing
	server.Router().GlobalMiddleware(&parse.Middleware{
		MaxUploadSize: 10,
	})

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
	taskSvc := taskService.NewService(taskRepository, presetSvc, webhookSvc, websocketSvc, ffmpegSvc).ProcessQueue()
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

	server.RegisterRoutes(func(_ *goyave.Server, router *goyave.Router) {
		router.Middleware(
			&middleware.CompressMiddleware{},
			&middleware.DebugoMiddleware{},
			&middleware.VersionMiddleware{},
		)

		apiRouter := router.Subrouter("/api/v1")
		apiRouter.Controller(&preset.Controller{})
		apiRouter.Controller(&version.Controller{})
		apiRouter.Controller(&preset.Controller{})
		apiRouter.Controller(&webhook.Controller{})
		apiRouter.Controller(&watchfolder.Controller{})
		apiRouter.Controller(&client.Controller{})
		apiRouter.Controller(&settings.Controller{})
		apiRouter.Controller(&task.Controller{})
		apiRouter.Controller(&debug.Controller{})

		// health
		router.Controller(&health.Controller{})

		// websocket
		apiRouter.Subrouter("/ws").Controller(ws.New(&websocket.Controller{}))

		// ui
		router.Controller(&ui.Controller{})

		// umami
		router.Controller(&umami.Controller{})

		// swagger
		router.Controller(&swagger.Controller{})

		// prometheus
		router.Controller(&prometheus.Controller{})

		// umami
		router.Controller(&umami.Controller{})
	})

	return server
}
