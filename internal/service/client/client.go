package client

import (
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/cfg"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/debug"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/internal/service/websocket"
)

type Repository interface {
	List(int, int) (*[]model.Client, int64, error)
	Add(*model.Client) (*model.Client, error)
	First() (*model.Client, error)
}

type Service struct {
	repository       Repository
	version          string
	websocketService *websocket.Service
}

func NewService(repository *repository.Client, version string, websocketService *websocket.Service) *Service {
	s := &Service{
		repository:       repository,
		version:          version,
		websocketService: websocketService,
	}

	// periodically update client info
	s.UpdateClientInfo()

	return s
}

func (s *Service) List(page int, perPage int) (*[]model.Client, int64, error) {
	return s.repository.List(page, perPage)
}

func (s *Service) Client() *dto.NewClient {
	return localClient
}

func (s *Service) save(newClient *dto.NewClient) (*model.Client, error) {
	c := &model.Client{
		Identifier: newClient.Identifier,
		Session:    newClient.Session,
		Cluster:    newClient.Cluster,

		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  s.version,
		FFMpeg:   cfg.GetString("ffmate.ffmpeg"),
		LastSeen: time.Now().UnixMilli(),
	}

	nc, err := s.repository.Add(c)
	s.websocketService.Broadcast(websocket.CLIENT_UPDATED, nc.ToDto())

	return nc, err
}

func (s *Service) first() (*model.Client, error) {
	return s.repository.First()
}

var localClient *dto.NewClient

func (s *Service) UpdateClientInfo() {
	localClient = &dto.NewClient{
		Identifier: cfg.GetString("ffmate.identifier"),
		Session:    cfg.GetString("ffmate.session"),
		Cluster:    uuid.NewString(),
	}
	debug.Client.Debug("initiated client as '%s'", localClient.Identifier)

	first, err := s.first()
	if err == nil && first != nil {
		localClient.Cluster = first.Cluster
	}

	// save client directly
	s.saveClient(localClient)

	// re-save client periodically
	go func() {
		for {
			time.Sleep(15 * time.Second)
			s.saveClient(localClient)

		}
	}()
}

func (s *Service) saveClient(c *dto.NewClient) {
	_, err := s.save(c)
	if err != nil {
		panic(err)
	}
	debug.Client.Debug("client info updated")
}

func (s *Service) Name() string {
	return service.Client
}
