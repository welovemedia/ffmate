package cmd

import (
	"fmt"
	"os"

	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	updateSvc "github.com/welovemedia/ffmate/v2/internal/service/update"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
)

var updater *selfupdate.Updater

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update ffmate",
	Run:   update,
}

var dry bool

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&dry, "dry", false, "run in dry mode (no real update)")
	viper.BindPFlag("dry", updateCmd.Flags().Lookup("dry"))

	updater = &selfupdate.Updater{
		CurrentVersion: viper.GetString("app.version"),
		ApiURL:         "https://earth.ffmate.io/_update/",
		BinURL:         "https://earth.ffmate.io/_update/",
		ForceCheck:     true,
		CmdName:        "ffmate",
	}
}

func update(cmd *cobra.Command, args []string) {
	server, err := goyave.New(goyave.Options{
		Config: config.LoadDefault(),
	})

	if err != nil {
		panic(err)
	}

	// register update service
	svc := updateSvc.NewService(viper.GetString("app.version"))
	server.RegisterService(svc)

	res, _, err := svc.CheckForUpdate(false, viper.GetBool("dry"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(res)
		os.Exit(0)
	}
}
