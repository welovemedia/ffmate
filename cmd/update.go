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

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("dry", false, "run in dry mode (no real update)")
	updateCmd.Flags().Bool("dev", false, "run in dev mode (check for dev updates)")
}

func update(cmd *cobra.Command, _ []string) {
	server, err := goyave.New(goyave.Options{
		Config: config.LoadDefault(),
	})

	if err != nil {
		debug.Log.Error("failed to initialize ffmate: %v", err)
		os.Exit(1)
	}

	svc := updateSvc.NewService(viper.GetString("app.version"))
	server.RegisterService(svc)

	dry, _ := cmd.Flags().GetBool("dry")
	dev, _ := cmd.Flags().GetBool("dev")

	res, _, err := svc.CheckForUpdate(false, dry, dev)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(res)
	os.Exit(0)
}
