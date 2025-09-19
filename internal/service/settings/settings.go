package settings

import (
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/service"
)

type Repository interface {
	Load() (*model.Settings, error)
	Store(*model.Settings) (*model.Settings, error)
}

type Service struct {
	repository Repository
}

func NewService(repository *repository.Settings) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Load() (*model.Settings, error) {
	return s.repository.Load()
}

func (s *Service) Store(newSetting *dto.Settings) (*model.Settings, error) {
	return s.repository.Store(&model.Settings{})
}

func (s *Service) Name() string {
	return service.Settings
}
