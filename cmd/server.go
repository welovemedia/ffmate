package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"fyne.io/systray"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/internal"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/utils"
	"github.com/welovemedia/ffmate/sev"
	"github.com/yosev/debugo"

	_ "embed"
)

//go:embed assets/icon_w.ico
var iconDataW []byte

//go:embed assets/icon.ico
var iconDataC []byte

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the server",
	Run:   start,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringP("ffmpeg", "f", "ffmpeg", "path to ffmpeg binary")
	serverCmd.PersistentFlags().StringP("port", "p", "3000", "the port to listen to")
	serverCmd.PersistentFlags().BoolP("tray", "t", false, "start with tray menu (experimental)")
	serverCmd.PersistentFlags().StringP("database", "b", "~/.ffmate/db.sqlite", "the path do the database")
	serverCmd.PersistentFlags().UintP("max-concurrent-tasks", "m", 3, "define maximum concurrent running tasks")

	serverCmd.PersistentFlags().BoolP("send-telemetry", "s", true, "enable sending anonymous telemetry data")

	viper.BindPFlag("ffmpeg", serverCmd.PersistentFlags().Lookup("ffmpeg"))
	viper.BindPFlag("port", serverCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("tray", serverCmd.PersistentFlags().Lookup("tray"))
	viper.BindPFlag("database", serverCmd.PersistentFlags().Lookup("database"))
	viper.BindPFlag("maxConcurrentTasks", serverCmd.PersistentFlags().Lookup("max-concurrent-tasks"))
	viper.BindPFlag("sendTelemetry", serverCmd.PersistentFlags().Lookup("send-telemetry"))
}

func start(cmd *cobra.Command, args []string) {
	config.Init()

	s := sev.New("ffmate", config.Config().AppVersion, config.Config().Database, config.Config().Port)
	switch config.Config().Loglevel {
	case "debug":
		s.Logger().SetLevel(logrus.DebugLevel)
	case "info":
		s.Logger().SetLevel(logrus.InfoLevel)
	case "warn":
		s.Logger().SetLevel(logrus.WarnLevel)
	case "error":
		s.Logger().SetLevel(logrus.ErrorLevel)
	case "fatal":
		s.Logger().SetLevel(logrus.FatalLevel)
	case "none":
		s.Logger().SetLevel(logrus.FatalLevel)
		s.Logger().SetOutput(io.Discard)
	}

	s.RegisterSignalHook()

	s.RegisterStartupHook(func(s *sev.Sev) {
		// broadcast all logs via websocket
		lb := &utils.LogBroadcaster{
			Callback: func(p []byte) {
				if service.WebsocketService() != nil {
					re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
					service.WebsocketService().Broadcast("log:created", re.ReplaceAllString(string(p), ""))
				}
			},
		}
		mw := io.MultiWriter(os.Stderr, lb)
		s.Logger().SetOutput(mw)
		debugo.SetOutput(mw)
	})

	s.RegisterStartupHook(func(s *sev.Sev) {
		s.Logger().Infof("server is listening on 0.0.0.0:%d", config.Config().Port)
	})
	if config.Config().SendTelemetry {
		_, err := os.Stat("/.dockerenv")
		s.RegisterShutdownHook(func(s *sev.Sev) {
			taskRepo := &repository.Task{DB: s.DB()}
			webhookRepo := &repository.Webhook{DB: s.DB()}
			presetRepo := &repository.Preset{DB: s.DB()}
			watchfolderRepo := &repository.Watchfolder{DB: s.DB()}
			count, _ := taskRepo.Count()
			countQueued, _ := taskRepo.CountByStatus(dto.QUEUED)
			countRunning, _ := taskRepo.CountByStatus(dto.RUNNING)
			countDoneSuccessful, _ := taskRepo.CountByStatus(dto.DONE_SUCCESSFUL)
			countDoneFailed, _ := taskRepo.CountByStatus(dto.DONE_ERROR)
			countDoneCanceled, _ := taskRepo.CountByStatus(dto.DONE_CANCELED)
			countWebhooks, _ := webhookRepo.Count()
			countPresets, _ := presetRepo.Count()
			countWatchfolder, _ := watchfolderRepo.Count()
			s.SendTelemetry(
				"https://telemetry.ffmate.io",
				map[string]interface{}{
					"Tasks":               count,
					"TasksQueued":         countQueued,
					"TasksRunning":        countRunning,
					"TasksDoneSuccessful": countDoneSuccessful,
					"TasksDoneFailed":     countDoneFailed,
					"TasksDoneCanceled":   countDoneCanceled,
					"Webhooks":            countWebhooks,
					"Presets":             countPresets,
					"Watchfolder":         countWatchfolder,
				},
				map[string]interface{}{
					"Tray":               config.Config().Tray,
					"Port":               config.Config().Port,
					"MaxConcurrentTasks": config.Config().MaxConcurrentTasks,
					"Debug":              config.Config().Debug,
					"Docker":             err == nil,
				},
			)
		})
	}

	internal.Init(s, config.Config().MaxConcurrentTasks, frontend)

	res, found, _ := updateAvailable()
	if found {
		s.Logger().Infof("found newer version %s (current: %s). Run '%s update' to update.", res, config.Config().AppVersion, config.Config().AppName)
	}

	readyFunc := func() {
		err := s.Start(config.Config().Port)
		if err != nil {
			s.Logger().Errorf("failed to start server: %s", err)
		}
	}

	if config.Config().Tray {
		useSystray(s, readyFunc)
	} else {
		readyFunc()
	}
}

func useSystray(s *sev.Sev, readyFunc func()) {

	s.RegisterShutdownHook(func(s *sev.Sev) {
		systray.Quit()
	})

	systray.Run(func() {
		systray.SetIcon(iconDataW)
		systray.SetTooltip(fmt.Sprintf("ffmate %s", config.Config().AppVersion))

		mFFmate := systray.AddMenuItem(fmt.Sprintf("ffmate %s", config.Config().AppVersion), "")
		mFFmate.SetIcon(iconDataC)
		mFFmate.Disable()

		systray.AddSeparator()

		mUi := systray.AddMenuItem("Open UI", "Open the ffmate ui")

		systray.AddSeparator()

		mQueued := systray.AddMenuItem("Queued tasks: 0", "")
		mQueued.Disable()
		mRunning := systray.AddMenuItem("Running tasks: 0", "")
		mRunning.Disable()
		mSuccessful := systray.AddMenuItem("Successful tasks: 0", "")
		mSuccessful.Disable()
		mError := systray.AddMenuItem("Failed tasks: 0", "")
		mError.Disable()
		mCanceled := systray.AddMenuItem("Canceled tasks: 0", "")
		mCanceled.Disable()

		systray.AddSeparator()
		res, found, _ := updateAvailable()
		mUpdate := systray.AddMenuItem("Check for updates", "Update ffmate")
		if found {
			mUpdate.SetTitle(fmt.Sprintf("Update available: %s", res))
		}
		mDebug := systray.AddMenuItemCheckbox("Enable debug", "Toggle debug", debugo.GetDebug() != "")

		systray.AddSeparator()

		mQuit := systray.AddMenuItem("Quit", "Quit ffmate")

		go func() {
			for {
				q, r, ds, de, dc, _ := service.TaskService().CountAllStatus(false)
				mQueued.SetTitle(fmt.Sprintf("Queued tasks: %d", q))
				mRunning.SetTitle(fmt.Sprintf("Running tasks: %d", r))
				mSuccessful.SetTitle(fmt.Sprintf("Successful tasks: %d", ds))
				mError.SetTitle(fmt.Sprintf("Failed tasks: %d", de))
				mCanceled.SetTitle(fmt.Sprintf("Canceled tasks: %d", dc))

				if r > 0 {
					systray.SetIcon(iconDataC)
				} else {
					systray.SetIcon(iconDataW)
				}

				time.Sleep(1 * time.Second)
			}
		}()

		go func() {
			for {
				select {
				case <-mUi.ClickedCh:
					url := fmt.Sprintf("http://localhost:%d", config.Config().Port)
					switch runtime.GOOS {
					case "linux":
						exec.Command("xdg-open", url).Start()
					case "darwin":
						exec.Command("open", url).Start()
					case "windows":
						exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
					}
				case <-mDebug.ClickedCh:
					if mDebug.Checked() {
						debugo.SetDebug("")
						mDebug.Uncheck()
					} else {
						debugo.SetDebug("*")
						mDebug.Check()
					}
				case <-mUpdate.ClickedCh:
					res, found, err := checkForUpdate(true)
					if err != nil {
						s.Logger().Error(err)
					} else {
						s.Logger().Info(res)
						if found {
							s.Logger().Info("please restart ffmate to apply the update")
							os.Exit(0)
						}
					}
				case <-mQuit.ClickedCh:
					s.Shutdown()
				}
			}

		}()
		readyFunc()
	}, func() {
	})
}
