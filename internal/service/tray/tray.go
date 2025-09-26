package tray

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/systray"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"github.com/welovemedia/ffmate/v2/internal/service/task"
	"github.com/welovemedia/ffmate/v2/internal/service/update"
	"github.com/yosev/debugo"
	"goyave.dev/goyave/v5"

	_ "embed"
)

//go:embed assets/icon_w.ico
var iconDataW []byte

//go:embed assets/icon.ico
var iconDataC []byte

type Service struct {
	server        *goyave.Server
	taskService   *task.Service
	updateService *update.Service
}

func NewService(server *goyave.Server, taskService *task.Service, updateService *update.Service) *Service {
	return &Service{
		server:        server,
		taskService:   taskService,
		updateService: updateService,
	}
}

func (s *Service) Run() {
	s.server.RegisterShutdownHook(func(_ *goyave.Server) {
		systray.Quit()
	})

	systray.Run(func() {
		if runtime.GOOS == "windows" {
			systray.SetIcon(iconDataC)
		} else {
			systray.SetIcon(iconDataW)
		}

		systray.SetTooltip(fmt.Sprintf("ffmate %s", s.server.Config().GetString("app.version")))

		mFFmate := systray.AddMenuItem(fmt.Sprintf("ffmate %s", s.server.Config().GetString("app.version")), "")
		mFFmate.SetIcon(iconDataC)
		mFFmate.Disable()

		systray.AddSeparator()

		mUI := systray.AddMenuItem("Open UI", "Open the ffmate ui")

		systray.AddSeparator()

		mQueued := systray.AddMenuItem("Queued tasks: 0", "")
		mQueued.Disable()
		mRunning := systray.AddMenuItem("Running tasks: 0", "")
		mRunning.Disable()
		mSuccessful := systray.AddMenuItem("Successful tasks: 0", "")
		mSuccessful.Disable()
		mError := systray.AddMenuItem("Failed tasks: 0", "")
		mError.Disable()
		mCanceled := systray.AddMenuItem("Canceled tasks: 0", "")
		mCanceled.Disable()

		systray.AddSeparator()
		res, found, _ := s.updateService.UpdateAvailable()
		mUpdate := systray.AddMenuItem("Check for updates", "Update ffmate")
		if found {
			mUpdate.SetTitle(fmt.Sprintf("Update available: %s", res))
		}
		mDebug := systray.AddMenuItemCheckbox("Enable debug", "Toggle debug", cfg.GetString("ffmate.debug") != "")

		systray.AddSeparator()

		mQuit := systray.AddMenuItem("Quit", "Quit ffmate")

		go func() {
			for {
				q, r, ds, de, dc, _ := s.taskService.CountAllStatus()
				mQueued.SetTitle(fmt.Sprintf("Queued tasks: %d", q))
				mRunning.SetTitle(fmt.Sprintf("Running tasks: %d", r))
				mSuccessful.SetTitle(fmt.Sprintf("Successful tasks: %d", ds))
				mError.SetTitle(fmt.Sprintf("Failed tasks: %d", de))
				mCanceled.SetTitle(fmt.Sprintf("Canceled tasks: %d", dc))

				if r > 0 {
					systray.SetIcon(iconDataC)
				} else {
					if runtime.GOOS == "windows" {
						systray.SetIcon(iconDataC)
					} else {
						systray.SetIcon(iconDataW)
					}
				}

				time.Sleep(1 * time.Second)
			}
		}()

		go func() {
			for {
				select {
				case <-mUI.ClickedCh:
					url := fmt.Sprintf("http://localhost:%d/ui", s.server.Config().GetInt("server.port"))
					switch runtime.GOOS {
					case "linux":
						_ = exec.Command("xdg-open", url).Start()
					case "darwin":
						_ = exec.Command("open", url).Start()
					case "windows":
						_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
					}
				case <-mDebug.ClickedCh:
					if mDebug.Checked() {
						debugo.SetNamespace("")
						mDebug.Uncheck()
					} else {
						debugo.SetNamespace("*")
						mDebug.Check()
					}
				case <-mUpdate.ClickedCh:
					res, found, err := s.updateService.CheckForUpdate(true, false)
					if err != nil {
						debug.Log.Error("%v", err)
					} else {
						debug.Log.Info("%s", res)
						if found {
							debug.Log.Info("please restart ffmate to apply the update")
							os.Exit(0)
						}
					}
				case <-mQuit.ClickedCh:
					s.server.Stop()
				}
			}
		}()
		if err := s.server.Start(); err != nil {
			debug.Log.Error("failed to start ffmate server: %v", err)
			os.Exit(1)
		}
	}, func() {
	})
}

func (s *Service) Name() string {
	return service.Tray
}
