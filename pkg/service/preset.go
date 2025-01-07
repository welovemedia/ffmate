package service

import (
	"errors"

	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/database/repository"
	"github.com/welovemedia/ffmate/pkg/dto"
	"github.com/welovemedia/ffmate/sev"
)

type PresetService struct {
	Sev              *sev.Sev
	PresetRepository *repository.Preset
	WebhookService   *WebhookService
}

func (s *PresetService) FindByName(name string) (*model.Preset, error) {
	w, err := s.PresetRepository.FirstByName(name)
	if err != nil {
		return nil, err
	}

	if w.Uuid == "" {
		return nil, errors.New("preset for given name not found")
	}

	return w, nil
}

func (s *PresetService) ListPresets() (*[]model.Preset, error) {
	return s.PresetRepository.List()
}

func (s *PresetService) DeletePreset(uuid string) error {
	w, err := s.PresetRepository.First(uuid)
	if err != nil {
		return err
	}

	if w.Uuid == "" {
		return errors.New("preset for given uuid not found")
	}

	err = s.PresetRepository.Delete(w)
	if err != nil {
		s.Sev.Logger().Warnf("failed to delete preset (uuid: %s): %+v", w.Uuid, err)
		return err
	}

	s.Sev.Logger().Infof("deleted preset (uuid: %s)", w.Uuid)

	s.Sev.Metrics().Gauge("preset.deleted").Inc()
	s.WebhookService.Fire(dto.PRESET_DELETE, w)

	return nil
}

func (s *PresetService) NewPreset(newPreset *dto.NewPreset) (*model.Preset, error) {
	w, err := s.PresetRepository.Create(newPreset.Command, newPreset.Name)
	s.Sev.Logger().Infof("created new preset (uuid: %s)", w.Uuid)

	s.Sev.Metrics().Gauge("preset.created").Inc()
	s.WebhookService.Fire(dto.PRESET_CREATED, w)

	return w, err
}
