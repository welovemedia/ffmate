package preset

import (
	"errors"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/metrics"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/service/webhook"
	"github.com/welovemedia/ffmate/internal/service/websocket"
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
		Uuid:           uuid.NewString(),
		Command:        newPreset.Command,
		Name:           newPreset.Name,
		Description:    newPreset.Description,
		Priority:       newPreset.Priority,
		Webhooks:       newPreset.Webhooks,
		Labels:         labels,
		OutputFile:     newPreset.OutputFile,
		PreProcessing:  newPreset.PreProcessing,
		PostProcessing: newPreset.PostProcessing,
	}
	w, err := s.repository.Save(preset)
	debug.Log.Info("created preset (uuid: %s)", w.Uuid)

	if newPreset.GlobalPresetName != "" {
		metrics.GaugeVec("preset.global").WithLabelValues(newPreset.GlobalPresetName).Inc()
	}

	metrics.Gauge("preset.created").Inc()
	s.webhookService.Fire(dto.PRESET_CREATED, w.ToDto())
	s.webhookService.FireDirect(w.Webhooks, dto.PRESET_CREATED, w.ToDto())
	s.websocketService.Broadcast(websocket.PRESET_CREATED, w.ToDto())

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
	w.PreProcessing = newPreset.PreProcessing
	w.PostProcessing = newPreset.PostProcessing
	w.OutputFile = newPreset.OutputFile
	w.Labels = labels
	w.Priority = newPreset.Priority
	w.Webhooks = newPreset.Webhooks

	w, err = s.repository.Save(w)
	if err != nil {
		debug.Log.Error("failed to update preset (uuid: %s): %v", w.Uuid, err)
		return nil, err
	}

	debug.Log.Error("updated preset (uuid: %s)", w.Uuid)

	if newPreset.GlobalPresetName != "" {
		metrics.GaugeVec("preset.global").WithLabelValues(newPreset.GlobalPresetName).Inc()
	}

	metrics.Gauge("preset.updated").Inc()
	s.webhookService.Fire(dto.PRESET_UPDATED, w.ToDto())
	s.webhookService.FireDirect(w.Webhooks, dto.PRESET_UPDATED, w.ToDto())
	s.websocketService.Broadcast(websocket.PRESET_UPDATED, w.ToDto())

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
		debug.Log.Error("failed to delete preset (uuid: %s)", uuid)
		return err
	}

	debug.Log.Info("deleted preset (uuid: %s)", uuid)

	metrics.Gauge("preset.deleted").Inc()
	s.webhookService.Fire(dto.PRESET_DELETED, w.ToDto())
	s.webhookService.FireDirect(w.Webhooks, dto.PRESET_DELETED, w.ToDto())
	s.websocketService.Broadcast(websocket.PRESET_DELETED, w.ToDto())

	return nil
}

func (s *Service) Name() string {
	return service.Preset
}
