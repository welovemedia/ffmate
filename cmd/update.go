package cmd

import (
	"fmt"
	"os"

	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"github.com/welovemedia/ffmate/internal/config"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update ffmate",
	Run:   update,
}

var dry bool

func init() {
	updateCmd.PersistentFlags().BoolVarP(&dry, "dry", "", false, "run in dry mode (no real update)")

	updater = &selfupdate.Updater{
		CurrentVersion: config.Config().AppVersion,
		ApiURL:         "https://earth.ffmate.io/_update_test/",
		BinURL:         "https://earth.ffmate.io/_update_test/",
		ForceCheck:     true,
		CmdName:        "ffmate",
	}
	rootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) {
	res, _, err := checkForUpdate(false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(res)
		os.Exit(0)
	}
}

func checkForUpdate(force bool) (string, bool, error) {
	res, found, err := updateAvailable()
	if err != nil {
		return "", false, fmt.Errorf("failed to contact update server: %+v", err)
	}

	if !found {
		return "no newer version found", false, nil
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

func updateAvailable() (string, bool, error) {
	res, err := updater.UpdateAvailable()
	if err != nil {
		return "", false, err
	}
	if res == "" || res == config.Config().AppVersion {
		return "", false, nil
	}

	return res, true, nil
}
