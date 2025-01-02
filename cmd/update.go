package cmd

import (
	"fmt"

	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/spf13/cobra"
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
		CurrentVersion: appVersion,
		ApiURL:         "https://ffmate.sev.wtf/_update/",
		BinURL:         "https://ffmate.sev.wtf/_update/",
		ForceCheck:     true,
		CmdName:        "ffmate",
	}

	rootCmd.AddCommand(updateCmd)
}

func updateAvailable() (string, bool, error) {
	res, err := updater.UpdateAvailable()
	if err != nil {
		return "", false, err
	}
	if res == "" || res == appVersion {
		return "", false, nil
	}

	return res, true, nil
}

func update(cmd *cobra.Command, args []string) {
	res, found, err := updateAvailable()
	if err != nil {
		fmt.Printf("failed to contact update server: %+v", err)
		return
	}

	if !found {
		fmt.Print("no newer version found\n")
		return
	}

	fmt.Printf("newer version found: %s\n", res)
	if !dry {
		err = updater.Update()
		if err != nil {
			fmt.Printf("failed to update to version:  %+v\n", err)
		} else {
			fmt.Printf("updated to version: %s\n", res)
		}
	}
}
