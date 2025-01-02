package cmd

import (
	"github.com/spf13/cobra"
	"github.com/welovemedia/ffmate/pkg"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/sev"
)

var serviceCmd = &cobra.Command{
	Use:   "server",
	Short: "start the server",
	Run:   start,
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}

func start(cmd *cobra.Command, args []string) {
	s := sev.New("ffmate", appVersion, dbPath, debug)

	s.RegisterSignalHook()

	s.RegisterStartupHook(func(s *sev.Sev) {
		s.Logger().Infof("server is listening on 0.0.0.0:%d", port)
	})
	if sendTelemetry {
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

	pkg.Init(s, concurrentTasks)

	res, found, _ := updateAvailable()
	if found {
		s.Logger().Infof("found newer version %s (current: %s). Run '%s update' to update.", res, s.AppVersion(), s.AppName())
	}

	err := s.Start(port)
	if err != nil {
		s.Logger().Errorf("failed to start server: %s", err)
	}
}
