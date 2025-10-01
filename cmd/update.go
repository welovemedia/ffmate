package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	updateSvc "github.com/welovemedia/ffmate/v2/internal/service/update"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update ffmate",
	Run:   update,
}

var dry bool

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&dry, "dry", false, "run in dry mode (no real update)")
	_ = viper.BindPFlag("dry", updateCmd.Flags().Lookup("dry"))
}

func update(_ *cobra.Command, _ []string) {
	server, err := goyave.New(goyave.Options{
		Config: config.LoadDefault(),
	})

	if err != nil {
		debug.Log.Error("failed to initialize ffmate: %v", err)
		os.Exit(1)
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
