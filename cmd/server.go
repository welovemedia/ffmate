package cmd

import (
	"fmt"

	"fyne.io/systray"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/internal"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
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
	serverCmd.PersistentFlags().StringP("port", "p", "3000", "the port to listen ob")
	serverCmd.PersistentFlags().BoolP("headless", "", false, "start without ui")
	serverCmd.PersistentFlags().BoolP("tray", "", false, "start with tray menu (experimental)")
	serverCmd.PersistentFlags().StringP("database", "", "db.sqlite", "the path do the database")
	serverCmd.PersistentFlags().UintP("max-concurrent-tasks", "m", 3, "define maximum concurrent running tasks")
	serverCmd.PersistentFlags().BoolP("send-telemetry", "", true, "enable sending anonymous telemetry data")

	viper.BindPFlag("ffmpeg", serverCmd.PersistentFlags().Lookup("ffmpeg"))
	viper.BindPFlag("port", serverCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("headless", serverCmd.PersistentFlags().Lookup("headless"))
	viper.BindPFlag("tray", serverCmd.PersistentFlags().Lookup("tray"))
	viper.BindPFlag("database", serverCmd.PersistentFlags().Lookup("database"))
	viper.BindPFlag("maxConcurrentTasks", serverCmd.PersistentFlags().Lookup("max-concurrent-tasks"))
	viper.BindPFlag("sendTelemetry", serverCmd.PersistentFlags().Lookup("send-telemetry"))
}

func start(cmd *cobra.Command, args []string) {
	config.Init()

	s := sev.New("ffmate", config.Config().AppVersion, config.Config().Database, config.Config().Port)

	s.RegisterSignalHook()

	s.RegisterStartupHook(func(s *sev.Sev) {
		s.Logger().Infof("server is listening on 0.0.0.0:%d", config.Config().Port)
	})
	if config.Config().SendTelemetry {
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
			s.SendTelemtry(
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
					"Headless":           config.Config().Headless,
					"Debug":              config.Config().Debug,
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
				select {
				case <-mDebug.ClickedCh:
					if mDebug.Checked() {
						debugo.SetDebug("")
						mDebug.Uncheck()
					} else {
						debugo.SetDebug("*")
						mDebug.Check()
					}
				case <-mUpdate.ClickedCh:
					checkForUpdate(true)
				case <-mQuit.ClickedCh:
					s.Shutdown()
				}
			}

		}()
		readyFunc()
	}, func() {
	})
}
