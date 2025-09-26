package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	promDto "github.com/prometheus/client_model/go"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
)

type Service struct {
	config *config.Config
	db     *gorm.DB
}

func NewService(config *config.Config, db *gorm.DB) *Service {
	return &Service{
		config: config,
		db:     db,
	}
}

func (s *Service) SendTelemetry(runtimeDuration time.Time, isShuttingDown bool, isStartUp bool) {
	b, err := json.Marshal(s.collectTelemetry(runtimeDuration, isShuttingDown, isStartUp))
	if err != nil {
		debug.Telemetry.Error("failed to marshal telemetry data: %+v", err)
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.GetString("ffmate.telemetry.url"), bytes.NewBuffer(b))
	if err != nil {
		debug.Telemetry.Error("failed to create http request: %+v", err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", s.config.GetString("app.name")+"/"+s.config.GetString("app.version"))

	resp, err := client.Do(req)
	if err != nil {
		debug.Telemetry.Error("failed to send telemetry data: %+v", err)
	}
	defer resp.Body.Close() // nolint:errcheck
}

func (s *Service) collectTelemetry(runtimeDuration time.Time, isShuttingDown bool, isStartUp bool) map[string]any {
	m := map[string]any{
		"appName":         s.config.GetString("app.name"),
		"appVersion":      s.config.GetString("app.version"),
		"isShuttingDown":  isShuttingDown,
		"isStartUp":       isStartUp,
		"runtimeDuration": time.Since(runtimeDuration).Milliseconds(),
		"os":              runtime.GOOS,
		"arch":            runtime.GOARCH,
		"session":         cfg.GetString("ffmate.session"),
		"config": map[string]any{
			"isTray":             cfg.GetBool("ffmate.isTray"),
			"maxConcurrentTasks": cfg.GetInt("ffmate.maxConcurrentTasks"),
			"debug":              cfg.GetString("ffmate.debug"),
			"isDocker":           cfg.GetBool("ffmate.isDocker"),
			"isCluster":          cfg.GetBool("ffmate.isCluster"),
			"port":               s.config.GetInt("server.port"),
		},
		"stats":   s.collectStats(),
		"metrics": s.getMetrics(),
	}

	if cfg.Has("ffmate.cluster") {
		m["cluster"] = cfg.GetString("ffmate.cluster")
	}

	return m
}

func (s *Service) collectStats() map[string]any {
	taskRepo := &repository.Task{DB: s.db}
	webhookRepo := &repository.Webhook{DB: s.db}
	presetRepo := &repository.Preset{DB: s.db}
	watchfolderRepo := &repository.Watchfolder{DB: s.db}

	count, _ := taskRepo.Count()
	countSourceWatchfolder, _ := taskRepo.CountAllBySource("watchfolder")
	countSourceAPI, _ := taskRepo.CountAllBySource("api")
	countDeleted, _ := taskRepo.CountDeleted()
	countQueued, _ := taskRepo.CountByStatus(dto.Queued)
	countRunning, _ := taskRepo.CountByStatus(dto.Running)
	countDoneSuccessful, _ := taskRepo.CountByStatus(dto.DoneSuccessful)
	countDoneFailed, _ := taskRepo.CountByStatus(dto.DoneError)
	countDoneCanceled, _ := taskRepo.CountByStatus(dto.DoneCanceled)
	countDeletedSuccessful, _ := taskRepo.CountDeletedByStatus(dto.DoneSuccessful)
	countDeletedFailed, _ := taskRepo.CountDeletedByStatus(dto.DoneError)
	countDeletedCanceled, _ := taskRepo.CountDeletedByStatus(dto.DoneCanceled)

	countWebhooks, _ := webhookRepo.Count()
	countWebhooksDeleted, _ := webhookRepo.CountDeleted()

	countPresets, _ := presetRepo.Count()
	countPresetsDeleted, _ := presetRepo.CountDeleted()

	countWatchfolder, _ := watchfolderRepo.Count()
	countWatchfolderDeleted, _ := watchfolderRepo.CountDeleted()

	return map[string]any{
		"tasks":                  count,
		"tasksDeleted":           countDeleted,
		"tasksQueued":            countQueued,
		"tasksRunning":           countRunning,
		"tasksDoneSuccessful":    countDoneSuccessful,
		"tasksDoneFailed":        countDoneFailed,
		"tasksDoneCanceled":      countDoneCanceled,
		"tasksDeletedSuccessful": countDeletedSuccessful,
		"tasksDeletedFailed":     countDeletedFailed,
		"tasksDeletedCanceled":   countDeletedCanceled,
		"taskSourceWatchfolder":  countSourceWatchfolder,
		"taskSourceAPI":          countSourceAPI,
		"webhooks":               countWebhooks,
		"webhooksDeleted":        countWebhooksDeleted,
		"presets":                countPresets,
		"presetsDeleted":         countPresetsDeleted,
		"watchfolder":            countWatchfolder,
		"watchfolderDeleted":     countWatchfolderDeleted,
	}
}

func (s *Service) getMetrics() map[string]float64 {
	var metricMap = make(map[string]float64)
	for name, gauge := range metrics.Gauges() {
		g := &promDto.Metric{}
		_ = gauge.Write(g)
		metricMap[name] = g.Gauge.GetValue()
	}
	for name, gaugeVec := range metrics.GaugesVec() {
		metricChan := make(chan prometheus.Metric, 1)

		go func() {
			gaugeVec.Collect(metricChan)
			close(metricChan)
		}()

		for metric := range metricChan {
			promMetric := &promDto.Metric{}
			if err := metric.Write(promMetric); err != nil {
				fmt.Printf("Error writing metric: %v\n", err)
				continue
			}

			labelValues := make([]string, len(promMetric.Label))
			for i, label := range promMetric.Label {
				labelValues[i] = fmt.Sprintf("%s=%s", *label.Name, *label.Value)
			}

			labeledName := fmt.Sprintf("%s{%s}", name, strings.Join(labelValues, ","))
			metricMap[labeledName] = promMetric.Gauge.GetValue()
		}
	}
	return metricMap
}

func (s *Service) Name() string {
	return service.Telemetry
}
