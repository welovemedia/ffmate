package client

import (
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/websocket"
)

type Repository interface {
	List(page int, perPage int) (*[]model.Client, int64, error)
	Add(client *model.Client) (*model.Client, error)
	First() (*model.Client, error)
}

type Service struct {
	repository       Repository
	websocketService *websocket.Service
	version          string
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
	s.websocketService.Broadcast(websocket.ClientUpdated, nc.ToDTO())

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

	// save for telemetry
	cfg.Set("ffmate.cluster", localClient.Cluster)

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
		debug.Log.Error("failed to save client: %v", err)
		return
	}
	debug.Client.Debug("client info updated")
}

func (s *Service) Name() string {
	return service.Client
}
