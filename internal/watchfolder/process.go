package watchfolder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/database/repository"
	"github.com/welovemedia/ffmate/internal/dto"
	"github.com/welovemedia/ffmate/internal/service"
	"github.com/welovemedia/ffmate/sev"
	"github.com/yosev/debugo"
)

type Watchfolder struct {
	Sev                   *sev.Sev
	WatchfolderRepository *repository.Watchfolder
}

var debug = debugo.New("watchfolder")

type fileState struct {
	Size     int64
	Attempts int
}

var (
	watchfolderCtx = sync.Map{}
)

func (w *Watchfolder) Init() {
	watchfolders, total, _ := w.WatchfolderRepository.List(-1, -1)
	debug.Debugf("initializing %d watchfolders", total)

	go w.monitorWatchfolderUpdates()

	for _, watchfolder := range *watchfolders {
		if !watchfolder.Suspended && !watchfolder.DeletedAt.Valid {
			ctx, cancel := context.WithCancelCause(context.Background())
			watchfolderCtx.Store(watchfolder.Uuid, cancel)
			go w.process(&watchfolder, ctx)
		}
	}
}

func (w *Watchfolder) monitorWatchfolderUpdates() {
	for {
		watchfolder := <-service.WatchfolderService().GetWatchfolderUpdates()
		if cancel, ok := watchfolderCtx.Load(watchfolder.Uuid); ok {
			// Cancel running watchfolder if found and remove context
			cancel.(context.CancelCauseFunc)(errors.New("updated"))
			watchfolderCtx.Delete(watchfolder.Uuid)
			debug.Debugf("canceled watchfolder (uuid: %s)", watchfolder.Uuid)
		}

		if !watchfolder.Suspended && !watchfolder.DeletedAt.Valid {
			ctx, cancel := context.WithCancelCause(context.Background())
			watchfolderCtx.Store(watchfolder.Uuid, cancel)
			// Create a deep copy of the watchfolder to avoid race conditions
			watchfolderCopy := *watchfolder // Create a copy of the struct
			go w.process(&watchfolderCopy, ctx)
		}
	}
}

func (w *Watchfolder) process(watchfolder *model.Watchfolder, ctx context.Context) {
	fileStates := sync.Map{}
	processedFiles := sync.Map{}
	debug.Debugf("initialized new watchfolder watcher (uuid: %s)", watchfolder.Uuid)

	for {
		select {
		case <-ctx.Done():
			w.Sev.Logger().Infof("stopped watchfolder (uuid: %s): %s", watchfolder.Uuid, context.Cause(ctx))
			return
		default:
		}
		debug.Debugf("processing watchfolder (uuid: %s)", watchfolder.Uuid)
		watchfolder.LastCheck = time.Now().UnixMilli()
		watchfolder.Error = ""

		// Walk the directory
		err := filepath.Walk(watchfolder.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Skip invisible files
			if strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}

			// Filter extensions
			if filterOutExtension(watchfolder, path) {
				return nil
			}

			// Check if the file has already been processed
			if _, seen := processedFiles.Load(path); seen {
				return nil
			}

			// Determine if the file is ready for processing
			if shouldProcessFile(path, info, &fileStates, watchfolder.GrowthChecks) {
				w.createTask(path, watchfolder)
				processedFiles.Store(path, true) // Mark as processed
				fileStates.Delete(path)          // Remove from tracking
			}

			return nil
		})

		if err != nil {
			watchfolder.Error = err.Error()
			w.Sev.Logger().Errorf("walking watchfolder directory failed (uuid: %s): %v", watchfolder.Uuid, err)
		}

		w.Sev.Metrics().Gauge("watchfolder.executed").Inc()
		service.WatchfolderService().UpdateWatchfolderInternal(watchfolder)
		time.Sleep(time.Duration(watchfolder.Interval * int(time.Second)))
	}
}

func (w *Watchfolder) createTask(path string, watchfolder *model.Watchfolder) {
	// create ffmate metadata map
	ffmate := map[string]map[string]string{
		"watchfolder": {
			"uuid": watchfolder.Uuid,
			"path": watchfolder.Path,
		},
	}
	metadata := &dto.InterfaceMap{"ffmate": ffmate}

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
		w.Sev.Logger().Errorf("failed to get relative path for %s (base: %s): %v", path, watchfolder.Path, err)
	} else {
		ffmate["watchfolder"]["relativePath"] = relPath
		ffmate["watchfolder"]["relativeDir"] = filepath.Dir(relPath)
	}

	// add new Task
	_, err = service.TaskService().NewTask(task, "", "watchfolder")
	if err != nil {
		w.Sev.Logger().Errorf("failed to create task for watchfolder (uuid: %s) file: %s: %v", watchfolder.Uuid, path, err)
		return
	}
	debug.Debugf("created new task for watchfolder (uuid: %s) file: %s", watchfolder.Uuid, path)
}

func filterOutExtension(watchfolder *model.Watchfolder, path string) bool {
	if watchfolder.Filter != nil && watchfolder.Filter.Extensions != nil {
		if len(watchfolder.Filter.Extensions.Exclude) > 0 {
			var exclude bool = false
			for _, ext := range watchfolder.Filter.Extensions.Exclude {
				if strings.HasSuffix(path, "."+ext) {
					exclude = true
					break
				}
			}
			return exclude
		}

		if len(watchfolder.Filter.Extensions.Include) > 0 {
			var include bool = true
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

// shouldProcessFile determines if a file is ready for processing based on growth attempts.
func shouldProcessFile(path string, info os.FileInfo, fileStates *sync.Map, growthChecks int) bool {
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
