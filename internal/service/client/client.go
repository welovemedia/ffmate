package client

import (
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/websocket"
)

type Repository interface {
	List(int, int) (*[]model.Client, int64, error)
	Save(*model.Client) (*model.Client, error)
	First() (*model.Client, error)
	Self(identifier string) (*model.Client, error)
}

type Service struct {
	repository       Repository
	websocketService *websocket.Service
	version          string
}

var identifier string

func NewService(repository *repository.Client, version string, websocketService *websocket.Service) *Service {
	s := &Service{
		repository:       repository,
		version:          version,
		websocketService: websocketService,
	}

	identifier = cfg.GetString("ffmate.identifier")

	// periodically update client info
	s.UpdateClientInfo(true)

	return s
}

func (s *Service) List(page int, perPage int) (*[]model.Client, int64, error) {
	return s.repository.List(page, perPage)
}

// hydrate newClient and safe to database
func (s *Service) save(newClient *model.Client) (*model.Client, error) {
	c := &model.Client{
		Identifier: newClient.Identifier,
		Session:    newClient.Session,
		Cluster:    newClient.Cluster,
		Labels:     newClient.Labels,

		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  s.version,
		FFMpeg:   cfg.GetString("ffmate.ffmpeg"),
		LastSeen: time.Now().UnixMilli(),
	}

	nc, err := s.repository.Save(c)
	s.websocketService.Broadcast(websocket.ClientUpdated, nc.ToDTO())

	return c, err
}

func (s *Service) first() (*model.Client, error) {
	return s.repository.First()
}

// Save client initially and start update loop
func (s *Service) UpdateClientInfo(startLoop bool) {
	var labels = make([]model.Label, len(cfg.GetStringSlice("ffmate.labels")))
	for i, label := range cfg.GetStringSlice("ffmate.labels") {
		labels[i] = model.Label{Value: label}
	}

	client := &model.Client{
		Identifier: cfg.GetString("ffmate.identifier"),
		Session:    cfg.GetString("ffmate.session"),
		Cluster:    uuid.NewString(),
		Labels:     labels,
	}
	debug.Client.Debug("initiated client as '%s'", client.Identifier)

	first, err := s.first()
	if err == nil && first != nil {
		client.Cluster = first.Cluster
	}

	// save client initially
	s.saveClient(client)

	// re-save client periodically
	if startLoop {
		go func() {
			for {
				time.Sleep(15 * time.Second)
				s.saveClient(client)

			}
		}()
	}
}

// save the client to the database
func (s *Service) saveClient(c *model.Client) {
	var err error
	_, err = s.save(c)
	if err != nil {
		panic(err)
	}
	debug.Client.Debug("client info updated")
}

func (s *Service) Name() string {
	return service.Client
}
