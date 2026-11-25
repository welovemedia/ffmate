package watchfolder

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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

	FirstAndLock(uuid string, eligible func(*model.Watchfolder) bool) (*model.Watchfolder, bool, error)
	UpdateScanResult(uuid string, lastCheck int64, scanError string) (*model.Watchfolder, error)
}

type Service struct {
	repository       Repository
	webhookService   *webhook.Service
	websocketService *websocket.Service
	taskService      *task.Service

	// running holds the uuids of all watchfolders this instance has a processing loop for
	running sync.Map
}

// clusterSyncInterval defines how often an instance in cluster mode checks the database
// for watchfolders it has no processing loop for yet (e.g. created on another instance)
var clusterSyncInterval = 5 * time.Second

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

	if err == nil {
		s.startWatchfolder(w)
	}

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
		debug.Log.Error("failed to list watchfolder: %v", err)
		os.Exit(1)
	}

	for _, watchfolder := range *watchfolders {
		s.startWatchfolder(&watchfolder)
	}

	// in cluster mode watchfolders may be created by other instances,
	// pick them up so every instance processes every watchfolder
	if cfg.GetBool("ffmate.isCluster") {
		go func() {
			for {
				time.Sleep(clusterSyncInterval)
				s.syncWatchfolders()
			}
		}()
	}
}

// syncWatchfolders starts a processing loop for every watchfolder in the database that has none on this instance yet
func (s *Service) syncWatchfolders() {
	watchfolders, _, err := s.List(-1, -1)
	if err != nil {
		debug.Watchfolder.Error("failed to list watchfolders for cluster sync: %v", err)
		return
	}
	for _, watchfolder := range *watchfolders {
		s.startWatchfolder(&watchfolder)
	}
}

// startWatchfolder starts the processing loop for the given watchfolder, unless one is already running on this instance
func (s *Service) startWatchfolder(watchfolder *model.Watchfolder) {
	if _, loaded := s.running.LoadOrStore(watchfolder.UUID, struct{}{}); loaded {
		return
	}
	debug.Watchfolder.Debug("watchfolder processing started (uuid: %s)", watchfolder.UUID)
	go func() {
		defer s.running.Delete(watchfolder.UUID)
		s.processWatchfolder(watchfolder)
	}()
}

// canProcess reports whether this instance should process the watchfolder
func canProcess(wf *model.Watchfolder) bool {
	return !wf.Suspended && hasLabelOverlap(wf.Labels, cfg.GetStringSlice("ffmate.labels"))
}

func (s *Service) processWatchfolder(watchfolder *model.Watchfolder) {
	fileStates := sync.Map{}
	isCluster := cfg.GetBool("ffmate.isCluster")
	// interval holds the most recently loaded value, so updates take effect from the next iteration on
	interval := watchfolder.Interval
	for {
		duration := time.Duration(max(interval, 1)) * time.Second
		next := time.Now().Truncate(duration).Add(duration)
		time.Sleep(time.Until(next))

		var wf *model.Watchfolder
		var locked = false
		var err error
		if isCluster {
			wf, locked, err = s.repository.FirstAndLock(watchfolder.UUID, canProcess)
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
		interval = wf.Interval
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

		err = s.scanWatchfolder(wf, &fileStates)

		// persist the result of this scan, a successful scan clears a previous error
		var scanError string
		if err != nil {
			scanError = err.Error()
			debug.Watchfolder.Error("walking watchfolder directory failed (uuid: %s): %v", watchfolder.UUID, err)
		}

		metrics.Gauge("watchfolder.executed").Inc()
		updated, err := s.repository.UpdateScanResult(wf.UUID, time.Now().UnixMilli(), scanError)
		if err != nil {
			debug.Log.Error("failed to update watchfolder internally (uuid: %s): %v", watchfolder.UUID, err)
		} else if updated != nil {
			s.websocketService.Broadcast(websocket.WatchfolderUpdated, updated.ToDTO())
		}
	}
}

// scanWatchfolder walks the watchfolder directory recursively and creates a task (and a .lock file)
// for every file that is ready for processing
func (s *Service) scanWatchfolder(wf *model.Watchfolder, fileStates *sync.Map) error {
	return filepath.Walk(wf.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip directories, invisible files, .lock files and filtered extensions
		base := filepath.Base(path)
		if info.IsDir() || strings.HasPrefix(base, ".") || strings.HasSuffix(base, ".lock") || s.filterOutExtension(wf, path) {
			return nil
		}

		// if a .lock file exists, the file has already been processed
		if _, err := os.Stat(path + ".lock"); err == nil {
			return nil
		}

		// determine if the file is ready for processing
		if !s.shouldProcessFile(path, info, fileStates, wf.GrowthChecks) {
			return nil
		}
		fileStates.Delete(path) // Remove from tracking

		// only lock the file if a task has actually been created, otherwise it is retried on the next interval
		if err := s.createTask(path, wf); err != nil {
			return nil
		}
		if err := os.WriteFile(path+".lock", []byte(""), 0777); err != nil {
			debug.Log.Error("failed to write .lock file (uuid: %s): %v", wf.UUID, err)
		}
		return nil
	})
}

// hasLabelOverlap checks if any task label matches any client label.
// Task labels (a) can contain '*' wildcards, which are converted to regex '.*'.
// Client labels (b) are literal strings.
func hasLabelOverlap(a []model.Label, b []string) bool {
	if len(a) == 0 {
		return true
	} else if len(b) == 0 {
		return false
	}

	for _, taskLabel := range a {
		pattern := taskLabel.Value

		regexPattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*") + "$"
		re, err := regexp.Compile(regexPattern)
		if err != nil {
			continue
		}

		if slices.ContainsFunc(b, func(clientLabel string) bool {
			return re.MatchString(clientLabel)
		}) {
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

	// Get or initialize the file state; the first sighting counts as the first check
	state, loaded := fileStates.LoadOrStore(path, &fileState{Size: info.Size(), Attempts: 1})
	fileState := state.(*fileState)

	if loaded {
		if info.Size() == fileState.Size {
			// size is stable, count this check
			fileState.Attempts++
		} else {
			// size changed, restart counting with this check
			fileState.Size = info.Size()
			fileState.Attempts = 1
		}
	}

	// ready once the size has been the same for growthChecks consecutive checks
	return fileState.Attempts >= growthChecks
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

func (s *Service) createTask(path string, watchfolder *model.Watchfolder) error {
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
		return err
	}
	debug.Watchfolder.Debug("created new task for watchfolder (uuid: %s), file: '%s'", watchfolder.UUID, path)
	return nil
}

func (s *Service) Name() string {
	return service.Watchfolder
}
