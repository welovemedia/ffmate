package preset

import (
	"errors"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/webhook"
	"github.com/welovemedia/ffmate/v2/internal/service/websocket"
)

type Repository interface {
	List(page int, perPage int) (*[]model.Preset, int64, error)
	Save(*model.Preset) (*model.Preset, error)
	First(string) (*model.Preset, error)
	Delete(*model.Preset) error
	Count() (int64, error)
}

type Service struct {
	repository       Repository
	webhookService   *webhook.Service
	websocketService *websocket.Service
}

func NewService(repository Repository, webhookService *webhook.Service, websocketService *websocket.Service) *Service {
	return &Service{
		repository:       repository,
		webhookService:   webhookService,
		websocketService: websocketService,
	}
}

func (s *Service) Get(uuid string) (*model.Preset, error) {
	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("preset for given uuid not found")
	}

	return w, nil
}

func (s *Service) List(page int, perPage int) (*[]model.Preset, int64, error) {
	return s.repository.List(page, perPage)
}

func (s *Service) Add(newPreset *dto.NewPreset) (*model.Preset, error) {
	var labels = make([]model.Label, len(newPreset.Labels))
	for i, label := range newPreset.Labels {
		labels[i] = model.Label{Value: label}
	}

	preset := &model.Preset{
		UUID:           uuid.NewString(),
		Command:        newPreset.Command,
		Name:           newPreset.Name,
		Description:    newPreset.Description,
		Retries:        newPreset.Retries,
		Priority:       newPreset.Priority,
		Webhooks:       newPreset.Webhooks,
		Labels:         labels,
		OutputFile:     newPreset.OutputFile,
		PreProcessing:  newPreset.PreProcessing,
		PostProcessing: newPreset.PostProcessing,
	}
	w, err := s.repository.Save(preset)
	debug.Preset.Info("created preset (uuid: %s)", w.UUID)

	if newPreset.GlobalPresetName != "" {
		metrics.GaugeVec("preset.global").WithLabelValues(newPreset.GlobalPresetName).Inc()
	}

	metrics.Gauge("preset.created").Inc()
	s.webhookService.Fire(dto.PresetCreated, w.ToDTO())
	s.webhookService.FireDirect(w.Webhooks, dto.PresetCreated, w.ToDTO())
	s.websocketService.Broadcast(websocket.PresetCreated, w.ToDTO())

	return w, err
}

func (s *Service) Update(uuid string, newPreset *dto.NewPreset) (*model.Preset, error) {
	var labels = make([]model.Label, len(newPreset.Labels))
	for i, label := range newPreset.Labels {
		labels[i] = model.Label{Value: label}
	}

	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("preset for given uuid not found")
	}

	w.Name = newPreset.Name
	w.Description = newPreset.Description
	w.Command = newPreset.Command
	w.Retries = newPreset.Retries
	w.PreProcessing = newPreset.PreProcessing
	w.PostProcessing = newPreset.PostProcessing
	w.OutputFile = newPreset.OutputFile
	w.Labels = labels
	w.Priority = newPreset.Priority
	w.Webhooks = newPreset.Webhooks

	w, err = s.repository.Save(w)
	if err != nil {
		debug.Preset.Error("failed to update preset (uuid: %s): %v", w.UUID, err)
		return nil, err
	}

	debug.Preset.Error("updated preset (uuid: %s)", w.UUID)

	if newPreset.GlobalPresetName != "" {
		metrics.GaugeVec("preset.global").WithLabelValues(newPreset.GlobalPresetName).Inc()
	}

	metrics.Gauge("preset.updated").Inc()
	s.webhookService.Fire(dto.PresetUpdated, w.ToDTO())
	s.webhookService.FireDirect(w.Webhooks, dto.PresetUpdated, w.ToDTO())
	s.websocketService.Broadcast(websocket.PresetUpdated, w.ToDTO())

	return w, err
}

func (s *Service) Delete(uuid string) error {
	w, err := s.repository.First(uuid)
	if err != nil {
		return err
	}

	if w == nil {
		return errors.New("preset for given uuid not found")
	}

	err = s.repository.Delete(w)
	if err != nil {
		debug.Preset.Error("failed to delete preset (uuid: %s)", uuid)
		return err
	}

	debug.Preset.Info("deleted preset (uuid: %s)", uuid)

	metrics.Gauge("preset.deleted").Inc()
	s.webhookService.Fire(dto.PresetDeleted, w.ToDTO())
	s.webhookService.FireDirect(w.Webhooks, dto.PresetDeleted, w.ToDTO())
	s.websocketService.Broadcast(websocket.PresetDeleted, w.ToDTO())

	return nil
}

func (s *Service) Name() string {
	return service.Preset
}
