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

func init() {
	rootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) {
	var updater = &selfupdate.Updater{
		CurrentVersion: appVersion,
		ApiURL:         "https://ffmate.sev.wtf/_update/",
		BinURL:         "https://ffmate.sev.wtf/_update/",
		ForceCheck:     true,
		CmdName:        "ffmate",
	}

	res, err := updater.UpdateAvailable()
	if err != nil {
		fmt.Printf("failed to contact update server: %+v", err)
	} else {
		if res == "" {
			fmt.Print("no newer version found\n")
			return
		}
		fmt.Printf("newer version found: %s\n", res)
		err = updater.Update()
		if err != nil {
			fmt.Printf("failed to update to version:  %+v\n", err)
		} else {
			fmt.Printf("updated to version: %s\n", res)
		}
	}
}
