package watchfolder

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/task"
	"github.com/welovemedia/ffmate/v2/internal/service/webhook"
	"github.com/welovemedia/ffmate/v2/internal/service/websocket"
)

type Repository interface {
	List(page int, perPage int) (*[]model.Watchfolder, int64, error)
	Save(watchfolder *model.Watchfolder) (*model.Watchfolder, error)
	First(uuid string) (*model.Watchfolder, error)
	Delete(watchfolder *model.Watchfolder) error
	Count() (int64, error)

	FirstAndLock(uuid string) (*model.Watchfolder, bool, error)
}

type Service struct {
	repository       Repository
	webhookService   *webhook.Service
	websocketService *websocket.Service
	taskService      *task.Service
}

func NewService(repository Repository, webhookService *webhook.Service, websocketService *websocket.Service, taskService *task.Service) *Service {
	return &Service{
		repository:       repository,
		webhookService:   webhookService,
		websocketService: websocketService,
		taskService:      taskService,
	}
}

func (s *Service) Get(uuid string) (*model.Watchfolder, error) {
	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("watchfolder for given uuid not found")
	}

	return w, nil
}

func (s *Service) List(page int, perPage int) (*[]model.Watchfolder, int64, error) {
	return s.repository.List(page, perPage)
}

func (s *Service) Add(newWatchfolder *dto.NewWatchfolder) (*model.Watchfolder, error) {
	var labels = make([]model.Label, len(newWatchfolder.Labels))
	for i, label := range newWatchfolder.Labels {
		labels[i] = model.Label{Value: label}
	}

	w, err := s.repository.Save(&model.Watchfolder{
		UUID: uuid.NewString(),

		Name:        newWatchfolder.Name,
		Description: newWatchfolder.Description,

		Path:         newWatchfolder.Path,
		Interval:     newWatchfolder.Interval,
		GrowthChecks: newWatchfolder.GrowthChecks,

		Filter: newWatchfolder.Filter,

		Preset: newWatchfolder.Preset,
		Labels: labels,

		Suspended: newWatchfolder.Suspended,
	})
	debug.Watchfolder.Info("created watchfolder (uuid: %s)", w.UUID)

	go s.processWatchfolder(w)

	metrics.Gauge("watchfolder.created").Inc()
	s.webhookService.Fire(dto.WatchfolderCreated, w.ToDTO())
	s.websocketService.Broadcast(websocket.WatchfolderCreated, w.ToDTO())

	return w, err
}

func (s *Service) Update(uuid string, newWatchfolder *dto.NewWatchfolder) (*model.Watchfolder, error) {
	var labels = make([]model.Label, len(newWatchfolder.Labels))
	for i, label := range newWatchfolder.Labels {
		labels[i] = model.Label{Value: label}
	}

	w, err := s.repository.First(uuid)
	if err != nil {
		return nil, err
	}

	if w == nil {
		return nil, errors.New("watchfolder for given uuid not found")
	}

	w.Name = newWatchfolder.Name
	w.Description = newWatchfolder.Description
	w.Path = newWatchfolder.Path
	w.Interval = newWatchfolder.Interval
	w.GrowthChecks = newWatchfolder.GrowthChecks
	w.Filter = newWatchfolder.Filter
	w.Preset = newWatchfolder.Preset
	w.Labels = labels
	w.Suspended = newWatchfolder.Suspended

	w, err = s.repository.Save(w)
	if err != nil {
		debug.Log.Error("failed to update watchfolder (uuid: %s): %v", w.UUID, err)
		return nil, err
	}

	debug.Watchfolder.Info("updated watchfolder (uuid: %s)", w.UUID)

	metrics.Gauge("watchfolder.updated").Inc()
	s.webhookService.Fire(dto.WatchfolderUpdated, w.ToDTO())
	s.websocketService.Broadcast(websocket.WatchfolderUpdated, w.ToDTO())

	return w, err
}

// UpdateInternal updates the whole watchfolder without previous validation (internal usage)
func (s *Service) UpdateInternal(watchfolder *model.Watchfolder) (*model.Watchfolder, error) {
	w, err := s.repository.Save(watchfolder)
	if err == nil {
		s.websocketService.Broadcast(websocket.WatchfolderUpdated, w.ToDTO())
	}
	return w, err
}

func (s *Service) Delete(uuid string) error {
	w, err := s.repository.First(uuid)
	if err != nil {
		return err
	}

	if w == nil {
		return errors.New("watchfolder for given uuid not found")
	}

	err = s.repository.Delete(w)
	if err != nil {
		debug.Log.Error("failed to delete watchfolder (uuid: %s)", uuid)
		return err
	}

	debug.Watchfolder.Info("deleted watchfolder (uuid: %s)", uuid)

	metrics.Gauge("watchfolder.deleted").Inc()
	s.webhookService.Fire(dto.WatchfolderDeleted, w.ToDTO())
	s.websocketService.Broadcast(websocket.WatchfolderDeleted, w.ToDTO())

	return nil
}

/**
 * watchfolder processing
 */

type fileState struct {
	Size     int64
	Attempts int
}

func (s *Service) Process() {
	watchfolders, _, err := s.List(-1, -1)
	if err != nil {
		panic(err)
	}

	for _, watchfolder := range *watchfolders {
		go s.processWatchfolder(&watchfolder)
	}
}

func (s *Service) processWatchfolder(watchfolder *model.Watchfolder) {
	fileStates := sync.Map{}
	isCluster := cfg.GetBool("ffmate.isCluster")
	for {
		duration := time.Duration(watchfolder.Interval) * time.Second
		next := time.Now().Truncate(duration).Add(duration)
		time.Sleep(time.Until(next))

		var wf *model.Watchfolder
		var locked = false
		var err error
		if isCluster {
			wf, locked, err = s.repository.FirstAndLock(watchfolder.UUID)
		} else {
			wf, err = s.repository.First(watchfolder.UUID)
		}
		if err != nil {
			debug.Log.Error("watchfolder requesting failed for processing: %v", err)
			continue
		}
		if wf == nil {
			debug.Watchfolder.Debug("watchfolder no longer exists - processing stopped (uuid: %s)", watchfolder.UUID)
			return
		}
		if wf.Suspended {
			debug.Watchfolder.Debug("watchfolder skipped as it is suspended (uuid: %s)", wf.UUID)
			continue
		}
		if !hasLabelOverlap(wf.Labels, cfg.GetStringSlice("ffmate.labels")) {
			debug.Watchfolder.Debug("watchfolder skipped as labels did not match (uuid: %s)", wf.UUID)
			continue
		}
		if locked {
			debug.Watchfolder.Debug("watchfolder skipped as it is locked (uuid: %s)", wf.UUID)
			continue
		}

		debug.Watchfolder.Debug("watchfolder processing (uuid: %s)", watchfolder.UUID)

		// walk the directory resursively
		err = filepath.Walk(wf.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// skip directories
			if info.IsDir() {
				return nil
			}

			// skip invisible files
			if strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}

			// skip .lock files
			if strings.HasSuffix(filepath.Base(path), ".lock") {
				return nil
			}

			// filter extensions
			if s.filterOutExtension(wf, path) {
				return nil
			}

			// if a .lock file exists, the file has already been processed
			if _, err := os.Stat(path + ".lock"); err == nil {
				return nil
			}

			// determine if the file is ready for processing
			if s.shouldProcessFile(path, info, &fileStates, wf.GrowthChecks) {
				fileStates.Delete(path) // Remove from tracking
				err := os.WriteFile(path+".lock", []byte(""), 0777)
				if err != nil {
					debug.Log.Error("failed to write .lock file (uuid: %s)", watchfolder.UUID)
				}
				s.createTask(path, wf)
			}

			return nil
		})

		if err != nil {
			watchfolder.Error = err.Error()
			debug.Watchfolder.Error("walking watchfolder directory failed (uuid: %s): %v", watchfolder.UUID, err)
		}

		metrics.Gauge("watchfolder.executed").Inc()
		wf.LastCheck = time.Now().UnixMilli()
		_, err = s.UpdateInternal(wf)
		if err != nil {
			debug.Log.Error("failed to update watchfolder internally (uuid: %s): %v", watchfolder.UUID, err)
		}
	}
}

// check if watchfolder and user labels have at least one overlap
func hasLabelOverlap(a []model.Label, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s.Value] = struct{}{}
	}

	for _, s := range b {
		if _, ok := set[s]; ok {
			return true
		}
	}
	return false
}

// shouldProcessFile determines if a file is ready for processing based on growth attempts.
func (s *Service) shouldProcessFile(path string, info os.FileInfo, fileStates *sync.Map, growthChecks int) bool {
	if growthChecks == 0 {
		// If no growth checks are required, the file is ready immediately
		return true
	}

	// Get or initialize the file state
	state, _ := fileStates.LoadOrStore(path, &fileState{Size: info.Size(), Attempts: 1})
	fileState := state.(*fileState)

	// Check if the file size is stable
	if info.Size() == fileState.Size {
		fileState.Attempts++
		if fileState.Attempts >= growthChecks {
			return true
		}
	} else {
		// File size changed, reset attempts
		fileState.Size = info.Size()
		fileState.Attempts = 1
		fileStates.Store(path, fileState)
	}

	return false
}

func (s *Service) filterOutExtension(watchfolder *model.Watchfolder, path string) bool {
	if watchfolder.Filter != nil && watchfolder.Filter.Extensions != nil {
		if len(watchfolder.Filter.Extensions.Exclude) > 0 {
			var exclude = false
			for _, ext := range watchfolder.Filter.Extensions.Exclude {
				if strings.HasSuffix(path, "."+ext) {
					exclude = true
					break
				}
			}
			return exclude
		}

		if len(watchfolder.Filter.Extensions.Include) > 0 {
			var include = true
			for _, ext := range watchfolder.Filter.Extensions.Include {
				if strings.HasSuffix(path, ext) {
					include = false
					break
				}
			}
			return include
		}
	}
	return false
}

func (s *Service) createTask(path string, watchfolder *model.Watchfolder) {
	// create ffmate metadata map
	ffmate := map[string]map[string]string{
		"watchfolder": {
			"uuid": watchfolder.UUID,
			"path": watchfolder.Path,
		},
	}
	metadata := &dto.MetadataMap{"ffmate": ffmate}

	// create new task
	task := &dto.NewTask{
		Preset:    watchfolder.Preset,
		Name:      filepath.Base(path),
		Metadata:  metadata,
		InputFile: path,
	}

	// hydrate paths in ffmate metadata map
	relPath, err := filepath.Rel(watchfolder.Path, path)
	if err != nil {
		debug.Log.Error("failed to get relative path for %s (base: %s): %v", path, watchfolder.Path, err)
	} else {
		ffmate["watchfolder"]["relativePath"] = relPath
		ffmate["watchfolder"]["relativeDir"] = filepath.Dir(relPath)
	}

	// add new Task
	_, err = s.taskService.Add(task, "watchfolder", "")
	if err != nil {
		debug.Log.Error("failed to create task for watchfolder (uuid: %s) file: %s: %v", watchfolder.UUID, path, err)
		return
	}
	debug.Watchfolder.Debug("created new task for watchfolder (uuid: %s), file: '%s'", watchfolder.UUID, path)
}

func (s *Service) Name() string {
	return service.Watchfolder
}
