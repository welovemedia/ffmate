package settings

import (
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/service"
)

type Repository interface {
	Load() (*model.Settings, error)
	Store(settings *model.Settings) (*model.Settings, error)
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

func (s *Service) Store(_ *dto.Settings) (*model.Settings, error) {
	return s.repository.Store(&model.Settings{})
}

func (s *Service) Name() string {
	return service.Settings
}
