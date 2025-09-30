package watchfolder

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/database/repository"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/service"
	clientService "github.com/welovemedia/ffmate/v2/internal/service/client"
	"github.com/welovemedia/ffmate/v2/internal/service/ffmpeg"
	presetService "github.com/welovemedia/ffmate/v2/internal/service/preset"
	settingsSvc "github.com/welovemedia/ffmate/v2/internal/service/settings"
	taskService "github.com/welovemedia/ffmate/v2/internal/service/task"
	"github.com/welovemedia/ffmate/v2/internal/service/telemetry"
	webhookService "github.com/welovemedia/ffmate/v2/internal/service/webhook"
	websocketService "github.com/welovemedia/ffmate/v2/internal/service/websocket"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/testutil"
)

func prepare(t *testing.T) (*testutil.TestServer, *Service) {
	server := testserver.New(t)
	// setup repositories
	presetRepository := (&repository.Preset{DB: server.DB()}).Setup()
	webhookRepository := (&repository.Webhook{DB: server.DB()}).Setup()
	webhookExecutionRepository := (&repository.WebhookExecution{DB: server.DB()}).Setup()
	watchfolderRepository := (&repository.Watchfolder{DB: server.DB()}).Setup()
	taskRepository := (&repository.Task{DB: server.DB()}).Setup()
	clientRepository := (&repository.Client{DB: server.DB()}).Setup()
	settingsRepository := (&repository.Settings{DB: server.DB()}).Setup()

	// setup and register services
	settingsService := settingsSvc.NewService(settingsRepository)
	telemetrySvc := telemetry.NewService(server.Config(), server.DB())
	ffmpegSvc := ffmpeg.NewService()
	websocketSvc := websocketService.NewService(server.DB())
	webhookSvc := webhookService.NewService(webhookRepository, webhookExecutionRepository, server.Config(), websocketSvc)
	presetSvc := presetService.NewService(presetRepository, webhookSvc, websocketSvc)
	taskSvc := taskService.NewService(taskRepository, presetSvc, webhookSvc, websocketSvc, ffmpegSvc, false)
	watchfolderSvc := NewService(watchfolderRepository, webhookSvc, websocketSvc, taskSvc)
	clientSvc := clientService.NewService(clientRepository, "test-1.0.0", websocketSvc)
	for _, svc := range map[string]goyave.Service{
		service.FFMpeg:      ffmpegSvc,
		service.Telemetry:   telemetrySvc,
		service.Websocket:   websocketSvc,
		service.Webhook:     webhookSvc,
		service.Preset:      presetSvc,
		service.Watchfolder: watchfolderSvc,
		service.Task:        taskSvc,
		service.Settings:    settingsService,
		service.Client:      clientSvc,
	} {
		server.RegisterService(svc)
	}

	return server, watchfolderSvc
}

func TestUpdateInternal(t *testing.T) {
	_, svc := prepare(t)
	m := &model.Watchfolder{
		UUID:        uuid.NewString(),
		Name:        "test",
		Description: "test",
	}
	w, err := svc.UpdateInternal(m)
	assert.NoError(t, err, "Successful update watchfolder")
	assert.False(t, w.Suspended, "Suspended should be false by default")

	// Toggle suspended and update again
	w.Suspended = true
	w, err = svc.UpdateInternal(w)
	assert.NoError(t, err, "Successful update watchfolder")
	assert.True(t, w.Suspended, "Suspended should be true after update")
}

func TestFilterOutExtension_Exclude(t *testing.T) {
	s := &Service{}
	wf := &model.Watchfolder{
		Filter: &dto.WatchfolderFilter{
			Extensions: &dto.WatchfolderFilterExtensions{
				Exclude: []string{"mp4"},
			},
		},
	}

	assert.True(t, s.filterOutExtension(wf, "/tmp/video.mp4"), "mp4 should be filtered out by exclude list")
	assert.False(t, s.filterOutExtension(wf, "/tmp/video.mov"), "mov should not be filtered out when only mp4 is excluded")
}

func TestFilterOutExtension_Include(t *testing.T) {
	s := &Service{}
	wf := &model.Watchfolder{
		Filter: &dto.WatchfolderFilter{
			Extensions: &dto.WatchfolderFilterExtensions{
				Include: []string{"mp4"},
			},
		},
	}

	assert.False(t, s.filterOutExtension(wf, "/tmp/video.mp4"), "mp4 should NOT be filtered out when included")
	assert.True(t, s.filterOutExtension(wf, "/tmp/video.mkv"), "mkv should be filtered out when not in include list")
}

type testFileInfo struct {
	name string
	size int64
	dir  bool
}

func (f testFileInfo) Name() string       { return f.name }
func (f testFileInfo) Size() int64        { return f.size }
func (f testFileInfo) Mode() os.FileMode  { return 0 }
func (f testFileInfo) ModTime() time.Time { return time.Now() }
func (f testFileInfo) IsDir() bool        { return f.dir }
func (f testFileInfo) Sys() any           { return nil }

func TestShouldProcessFile_NoGrowthChecks(t *testing.T) {
	s := &Service{}
	var states sync.Map
	info := testFileInfo{name: "a.mp4", size: 123}

	ok := s.shouldProcessFile("/tmp/a.mp4", info, &states, 0)
	assert.True(t, ok, "File should be processed immediately when growthChecks is 0")
}

func TestShouldProcessFile_StableSize(t *testing.T) {
	s := &Service{}
	var states = sync.Map{}
	path := "/tmp/a.mp4"

	// growthChecks=3 -> first pass false, second pass true (size stable)
	ok1 := s.shouldProcessFile(path, testFileInfo{name: "a.mp4", size: 100}, &states, 3)
	assert.False(t, ok1, "First pass should not process yet")

	ok2 := s.shouldProcessFile(path, testFileInfo{name: "a.mp4", size: 100}, &states, 3)
	assert.True(t, ok2, "Second pass with stable size should process")
}

func TestShouldProcessFile_SizeChangesResetsAttempts(t *testing.T) {
	s := &Service{}
	var states = sync.Map{}
	path := "/tmp/a.mp4"

	ok := s.shouldProcessFile(path, testFileInfo{name: "a.mp4", size: 100}, &states, 3)
	assert.False(t, ok, "First pass should not process yet")

	ok = s.shouldProcessFile(path, testFileInfo{name: "a.mp4", size: 200}, &states, 3)
	assert.False(t, ok, "Size changed: attempts reset, still not ready")

	ok = s.shouldProcessFile(path, testFileInfo{name: "a.mp4", size: 200}, &states, 3)
	assert.False(t, ok, "Second stable check after change should still not be ready with growthChecks=3")
	ok = s.shouldProcessFile(path, testFileInfo{name: "a.mp4", size: 200}, &states, 3)
	assert.True(t, ok, "Third stable check after change should process")
}

func TestCreateTask_CreatesTaskWithMetadata(t *testing.T) {
	server, svc := prepare(t)

	// Prepare a watchfolder and a nested file path
	base := t.TempDir()
	sub := filepath.Join(base, "incoming", "clips")
	requireNoErr := func(err error) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	requireNoErr(os.MkdirAll(sub, 0o755))

	wf := &model.Watchfolder{
		UUID: uuid.NewString(),
		Path: base,
		// Keep Preset empty to avoid preset resolution branch in task.Add
		Preset: "",
		Name:   "WF",
	}

	filePath := filepath.Join(sub, "video.mp4")

	// createTask should create a task synchronously
	svc.createTask(filePath, wf)

	// Query created tasks
	taskRepo := &repository.Task{DB: server.DB()}
	tasks, _, err := taskRepo.List(0, 10)
	assert.NoError(t, err, "listing tasks should not error")
	if assert.NotNil(t, tasks, "tasks should not be nil") && assert.GreaterOrEqual(t, len(*tasks), 1, "at least one task should exist") {
		created := (*tasks)[0]

		// Validate basic fields
		assert.Equal(t, "video.mp4", created.Name, "task name should be the file base")
		if assert.NotNil(t, created.InputFile, "InputFile should not be nil") {
			assert.Equal(t, filePath, created.InputFile.Raw, "InputFile.Raw should match the provided path")
		}
		// Source should be "watchfolder"
		assert.Equal(t, dto.TaskSource("watchfolder"), created.Source, "task source should be 'watchfolder'")

		// Validate ffmate metadata
		if assert.NotNil(t, created.Metadata, "metadata should not be nil") {
			root := (*created.Metadata)
			fv, ok := root["ffmate"]
			if assert.True(t, ok, "ffmate metadata should exist") {
				ffmateMap, ok := fv.(map[string]any)
				if assert.True(t, ok, "ffmate should be a map") {
					wfv, ok := ffmateMap["watchfolder"]
					if assert.True(t, ok, "watchfolder metadata should exist") {
						wfMap, ok := wfv.(map[string]any)
						if assert.True(t, ok, "watchfolder should be a map") {
							if got, ok := wfMap["uuid"].(string); assert.True(t, ok, "uuid should be string") {
								assert.Equal(t, wf.UUID, got, "watchfolder.uuid should match")
							}
							if got, ok := wfMap["path"].(string); assert.True(t, ok, "path should be string") {
								assert.Equal(t, wf.Path, got, "watchfolder.path should match")
							}
							// Relative path and dir
							rel, err := filepath.Rel(wf.Path, filePath)
							assert.NoError(t, err, "should compute relative path")
							if got, ok := wfMap["relativePath"].(string); assert.True(t, ok, "relativePath should be string") {
								assert.Equal(t, rel, got, "relativePath should match")
							}
							if got, ok := wfMap["relativeDir"].(string); assert.True(t, ok, "relativeDir should be string") {
								assert.Equal(t, filepath.Dir(rel), got, "relativeDir should match")
							}
						}
					}
				}
			}
		}
	}
}
