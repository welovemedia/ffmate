package internal

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/welovemedia/ffmate/internal/cfg"
	"github.com/welovemedia/ffmate/internal/controller"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/middleware"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/service/client"
	"github.com/welovemedia/ffmate/internal/service/ffmpeg"
	"github.com/welovemedia/ffmate/internal/service/preset"
	"github.com/welovemedia/ffmate/internal/service/settings"
	"github.com/welovemedia/ffmate/internal/service/task"
	"github.com/welovemedia/ffmate/internal/service/telemetry"
	"github.com/welovemedia/ffmate/internal/service/tray"
	"github.com/welovemedia/ffmate/internal/service/update"
	"github.com/welovemedia/ffmate/internal/service/watchfolder"
	"github.com/welovemedia/ffmate/internal/service/webhook"
	"github.com/welovemedia/ffmate/internal/service/websocket"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/cors"
	"goyave.dev/goyave/v5/middleware/parse"
)

func Init(options goyave.Options) {
	// create a new Goyave instance
	options.Logger = nil
	server, err := goyave.New(options)
	if err != nil {
		panic(err)
	}

	// setup cors
	server.Router().CORS(&cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
	})

	// add global parse middleware
	server.Router().GlobalMiddleware(&parse.Middleware{
		MaxUploadSize: 10,
	})

	// register startup hooks
	var startTime time.Time
	server.RegisterSignalHook()
	server.RegisterStartupHook(func(*goyave.Server) {
		debug.Log.Info("server started on port '%d' using driver '%s'", server.Config().GetInt("server.port"), server.Config().GetString("database.connection"))
		startTime = time.Now()

		// open the UI in the browser
		if cfg.GetBool("ffmate.isUI") && !cfg.GetBool("ffmate.isDocker") {
			url := fmt.Sprintf("http://localhost:%d/ui", server.Config().GetInt("server.port"))
			switch runtime.GOOS {
			case "linux":
				exec.Command("xdg-open", url).Start()
			case "darwin":
				exec.Command("open", url).Start()
			case "windows":
				exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			}
		}
	})

	// setup repositories
	presetRepository := (&repository.Preset{DB: server.DB()}).Setup()
	webhookRepository := (&repository.Webhook{DB: server.DB()}).Setup()
	webhookExecutionRepository := (&repository.WebhookExecution{DB: server.DB()}).Setup()
	watchfolderRepository := (&repository.Watchfolder{DB: server.DB()}).Setup()
	taskRepository := (&repository.Task{DB: server.DB()}).Setup()
	settingRepository := (&repository.Settings{DB: server.DB()}).Setup()
	clientRepository := (&repository.Client{DB: server.DB()}).Setup()

	// setup and register services
	telemetrySvc := telemetry.NewService(server.Config(), server.DB())
	ffmpegSvc := ffmpeg.NewService()
	updateSvc := update.NewService(server.Config().GetString("app.version"))
	websocketSvc := websocket.NewService(server.DB())
	webhookSvc := webhook.NewService(webhookRepository, webhookExecutionRepository, server.Config(), websocketSvc)
	presetSvc := preset.NewService(presetRepository, webhookSvc, websocketSvc)
	taskSvc := task.NewService(taskRepository, presetSvc, webhookSvc, websocketSvc, ffmpegSvc).ProcessQueue()
	traySvc := tray.NewService(server, taskSvc, updateSvc)
	watchfolderSvc := watchfolder.NewService(watchfolderRepository, webhookSvc, websocketSvc, taskSvc)
	settingSvc := settings.NewService(settingRepository)
	clientSvc := client.NewService(clientRepository, server.Config().GetString("app.version"), websocketSvc)
	for name, svc := range map[string]goyave.Service{
		service.Tray:        traySvc,
		service.Update:      updateSvc,
		service.FFMpeg:      ffmpegSvc,
		service.Telemetry:   telemetrySvc,
		service.Websocket:   websocketSvc,
		service.Webhook:     webhookSvc,
		service.Preset:      presetSvc,
		service.Watchfolder: watchfolderSvc,
		service.Task:        taskSvc,
		service.Settings:    settingSvc,
		service.Client:      clientSvc,
	} {
		server.RegisterService(svc)
		debug.Service.Debug("registered %s service", name)
	}

	// register routes
	server.RegisterRoutes(controller.Register)

	// register middlewares
	server.RegisterRoutes(middleware.Register)

	// setup debug logger to broadcast to websocket clients
	debug.RegisterBroadcastLogger(func(p []byte) {
		re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
		websocketSvc.Broadcast(websocket.LOG, re.ReplaceAllString(string(p), ""))
	})

	// init cluster if enabled
	if cfg.GetBool("ffmate.isCluster") {
		go websocketSvc.InitCluster()
	}

	// setup telemetry (if enabled)
	if cfg.GetBool("ffmate.telemetry.send") {
		// send telemetry data on shutdown
		server.RegisterShutdownHook(func(*goyave.Server) {
			telemetrySvc.SendTelemetry(startTime, true, false)
		})

		// send telemetry on startup and every 3 hours
		server.RegisterStartupHook(func(*goyave.Server) {
			var startup = true
			go func() {
				const interval = 3 * time.Hour
				for {
					telemetrySvc.SendTelemetry(startTime, false, startup)
					next := time.Now().Truncate(interval).Add(interval)
					startup = false
					time.Sleep(time.Until(next))
				}
			}()
		})
	}

	// start watchfolder processor
	watchfolderSvc.Process()

	// enable tray
	if cfg.GetBool("ffmate.isTray") {
		traySvc.Run()
	} else {
		if err := server.Start(); err != nil {
			panic(err)
		}
	}
}
