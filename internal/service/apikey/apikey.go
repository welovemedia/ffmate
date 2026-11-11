package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/service"
)

const (
	keyPrefix    = "ffm_"
	prefixLength = len(keyPrefix) + 8 // stored in clear to identify a key in listings
)

type Repository interface {
	List(page int, perPage int) (*[]model.APIKey, int64, error)
	Add(apiKey *model.APIKey) (*model.APIKey, error)
	First(uuid string) (*model.APIKey, error)
	FirstByHash(hash string) (*model.APIKey, error)
	Save(apiKey *model.APIKey) (*model.APIKey, error)
	Delete(apiKey *model.APIKey) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// Add creates a new api key and returns it together with the plain key.
// The plain key is not stored and cannot be retrieved again.
func (s *Service) Add(newAPIKey *dto.NewAPIKey) (*model.APIKey, string, error) {
	key, err := generateKey()
	if err != nil {
		return nil, "", err
	}

	a, err := s.repository.Add(&model.APIKey{
		UUID:    uuid.NewString(),
		Name:    newAPIKey.Name,
		Access:  newAPIKey.Access,
		Prefix:  key[:prefixLength],
		KeyHash: hashKey(key),
	})
	if err != nil {
		return nil, "", err
	}

	debug.Log.Info("created api key (uuid: %s)", a.UUID)
	return a, key, nil
}

// Rotate replaces the secret of an existing api key and returns it together with the new plain key.
// The api key keeps its uuid, name and access; the previous key is invalid immediately.
func (s *Service) Rotate(uuid string) (*model.APIKey, string, error) {
	a, err := s.Get(uuid)
	if err != nil {
		return nil, "", err
	}

	key, err := generateKey()
	if err != nil {
		return nil, "", err
	}

	a.Prefix = key[:prefixLength]
	a.KeyHash = hashKey(key)
	a, err = s.repository.Save(a)
	if err != nil {
		return nil, "", err
	}

	debug.Log.Info("rotated api key (uuid: %s)", a.UUID)
	return a, key, nil
}

// Resolve returns the api key matching the given plain key or nil if none matches
func (s *Service) Resolve(key string) (*model.APIKey, error) {
	if key == "" {
		return nil, nil
	}
	return s.repository.FirstByHash(hashKey(key))
}

func (s *Service) Get(uuid string) (*model.APIKey, error) {
	a, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("api key for given uuid not found")
	}
	return a, nil
}

func (s *Service) List(page int, perPage int) (*[]model.APIKey, int64, error) {
	return s.repository.List(page, perPage)
}

func (s *Service) Delete(uuid string) error {
	a, err := s.Get(uuid)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(a); err != nil {
		return err
	}

	debug.Log.Info("deleted api key (uuid: %s)", a.UUID)
	return nil
}

func (s *Service) Name() string {
	return service.APIKey
}

func generateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return keyPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// keys are high-entropy random values, so a fast unsalted hash is sufficient
// and allows an indexed lookup without a constant-time comparison
func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
