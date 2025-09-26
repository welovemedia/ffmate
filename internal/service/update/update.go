package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/v2/internal/service"
)

var updater = &selfupdate.Updater{
	CurrentVersion: viper.GetString("app.version"),
	ApiURL:         "https://earth.ffmate.io/_update/",
	BinURL:         "https://earth.ffmate.io/_update/",
	ForceCheck:     true,
	CmdName:        "ffmate",
}

type Service struct {
	version string
}

func NewService(version string) *Service {
	return &Service{
		version: version,
	}
}

func (s *Service) CheckForUpdate(force bool, dry bool) (string, bool, error) {
	res, found, err := s.UpdateAvailable()
	if err != nil {
		return "", false, fmt.Errorf("failed to contact update server: %+v", err)
	}

	if !found {
		return "no newer version found", false, nil
	}

	if dry && !force {
		return fmt.Sprintf("found newer version: %s\n", res), true, nil
	}

	if s.isHomebrew() {
		fmt.Println("homebrew installation detected - please use homebrew to update ffmate")
		os.Exit(0)
	}

	if !dry || force {
		err = updater.Update()
		if err != nil {
			return "", true, fmt.Errorf("failed to update to version: %+v", err)
		} else {
			return fmt.Sprintf("updated to version: %s\n", res), true, nil
		}
	}
	return "no updates found", false, nil
}

func (s *Service) UpdateAvailable() (string, bool, error) {
	res, err := updater.UpdateAvailable()
	if err != nil {
		return "", false, err
	}
	if res == "" || res == s.version {
		return "", false, nil
	}

	return res, true, nil
}

func (s *Service) isHomebrew() bool {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exePath, _ = filepath.EvalSymlinks(exePath)
	return strings.HasPrefix(exePath, "/opt/homebrew/")
}

func (s *Service) Name() string {
	return service.Update
}
