package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/pkg"
	"github.com/welovemedia/ffmate/pkg/config"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/sev"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the server",
	Run:   start,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringP("ffmpeg", "f", "ffmpeg", "path to ffmpeg binary")
	serverCmd.PersistentFlags().StringP("port", "p", "3000", "the port to listen ob")
	serverCmd.PersistentFlags().StringP("database", "", "db.sqlite", "the path do the database")
	serverCmd.PersistentFlags().UintP("max-concurrent-tasks", "m", 1, "define maximum concurrent running tasks")
	serverCmd.PersistentFlags().BoolP("send-telemetry", "", true, "enable sending anonymous telemetry data")

	viper.BindPFlag("ffmpeg", serverCmd.PersistentFlags().Lookup("ffmpeg"))
	viper.BindPFlag("port", serverCmd.PersistentFlags().Lookup("port"))
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
			count, _ := taskRepo.Count()
			countQueued, _ := taskRepo.CountByStatus(dto.QUEUED)
			countRunning, _ := taskRepo.CountByStatus(dto.RUNNING)
			countDoneSuccessful, _ := taskRepo.CountByStatus(dto.DONE_SUCCESSFUL)
			countDoneFailed, _ := taskRepo.CountByStatus(dto.DONE_ERROR)
			countDoneCanceled, _ := taskRepo.CountByStatus(dto.DONE_CANCELED)
			countWebhooks, _ := webhookRepo.Count()
			s.SendTelemtry(
				"https://eu-central-1.app.helmut.cloud/api/high5/v1/org/sev.wtf/spaces/Telegram/execute/webhook/804b722256a60614806444fee6859b8130bd652a7a6f589d887bdc3cfdf5de603a862f5f0e551ce4f4e581296a1a595dc5b286329cce2b2735b8657374c8d413",
				map[string]interface{}{
					"Tasks":               count,
					"TasksQueued":         countQueued,
					"TasksRunning":        countRunning,
					"TasksDoneSuccessful": countDoneSuccessful,
					"TasksDoneFailed":     countDoneFailed,
					"TasksDoneCanceled":   countDoneCanceled,
					"Webhooks":            countWebhooks,
				},
			)
		})
	}

	pkg.Init(s, config.Config().MaxConcurrentTasks)

	res, found, _ := updateAvailable()
	if found {
		s.Logger().Infof("found newer version %s (current: %s). Run '%s update' to update.", res, config.Config().AppVersion, config.Config().AppName)
	}

	err := s.Start(config.Config().Port)
	if err != nil {
		s.Logger().Errorf("failed to start server: %s", err)
	}
}
