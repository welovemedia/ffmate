package service

import (
	"errors"

	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/sev"
)

type watchfolderSvc struct {
	service
	sev                   *sev.Sev
	watchfolderRepository *repository.Watchfolder
}

var watchfolderUpdates = make(chan *model.Watchfolder, 100)

func (s *watchfolderSvc) GetWatchfolderUpdates() chan *model.Watchfolder {
	return watchfolderUpdates
}

func (s *watchfolderSvc) ListWatchfolders(page int, perPage int) (*[]model.Watchfolder, int64, error) {
	return s.watchfolderRepository.List(page, perPage)
}

func (s *watchfolderSvc) GetWatchfolderById(uuid string) (*model.Watchfolder, error) {
	return s.watchfolderRepository.First(uuid)
}

func (s *watchfolderSvc) UpdateWatchfolderInternal(watchfolder *model.Watchfolder) (*model.Watchfolder, error) {
	w, err := s.watchfolderRepository.Update(watchfolder)
	if err == nil {
		s.sev.Metrics().Gauge("watchfolder.updated").Inc()
		WebhookService().Fire(dto.WATCHFOLDER_UPDATED, w.ToDto())
		WebsocketService().Broadcast(WATCHFOLDER_UPDATED, w.ToDto())
	}
	return w, err
}

func (s *watchfolderSvc) DeleteWatchfolder(uuid string) error {
	w, err := s.watchfolderRepository.First(uuid)
	if err != nil {
		return err
	}

	if w.Uuid == "" {
		return errors.New("watchfolder for given uuid not found")
	}

	err = s.watchfolderRepository.Delete(w)
	if err != nil {
		s.sev.Logger().Warnf("failed to delete watchfolder (uuid: %s): %+v", w.Uuid, err)
		return err
	}

	s.sev.Logger().Infof("deleted watchfolder (uuid: %s)", w.Uuid)
	watchfolderUpdates <- w

	s.sev.Metrics().Gauge("watchfolder.deleted").Inc()
	WebhookService().Fire(dto.WATCHFOLDER_DELETED, w.ToDto())
	WebsocketService().Broadcast(WATCHFOLDER_DELETED, w.ToDto())

	return nil
}

func (s *watchfolderSvc) NewWatchfolder(newWatchfolder *dto.NewWatchfolder) (*model.Watchfolder, error) {
	_, err := PresetService().FindByUuid(newWatchfolder.Preset)
	if err != nil {
		return nil, err
	}
	w, err := s.watchfolderRepository.Create(newWatchfolder)

	s.sev.Logger().Infof("created new watchfolder (uuid: %s)", w.Uuid)
	watchfolderUpdates <- w

	s.sev.Metrics().Gauge("watchfolder.created").Inc()
	WebhookService().Fire(dto.WATCHFOLDER_CREATED, w.ToDto())
	WebsocketService().Broadcast(WATCHFOLDER_CREATED, w.ToDto())

	return w, err
}

func (s *watchfolderSvc) UpdateWatchfolder(watchfolderUuid string, newWatchfolder *dto.NewWatchfolder) (*model.Watchfolder, error) {
	w, err := s.GetWatchfolderById(watchfolderUuid)
	if err != nil {
		return nil, err
	}

	w.Name = newWatchfolder.Name
	w.Description = newWatchfolder.Description
	w.Path = newWatchfolder.Path
	w.Preset = newWatchfolder.Preset
	w.GrowthChecks = newWatchfolder.GrowthChecks
	w.Interval = newWatchfolder.Interval
	w.Filter = newWatchfolder.Filter

	watchfolderUpdates <- w

	return s.UpdateWatchfolderInternal(w)
}
