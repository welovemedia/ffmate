package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/yosev/debugo"
)

var updater *selfupdate.Updater
var frontend embed.FS

var rootCmd = &cobra.Command{
	Use:   "ffmate",
	Short: "ffmate is a wrapper for ffmpeg",
	Long:  "ffmate is a wrapper for ffmpeg that adds a queue system on top of it",
}

func init() {
	rootCmd.PersistentFlags().StringP("debug", "d", "", "set debugo namespace (eg. '*')")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func Execute(args []string, frontendFs embed.FS) {
	frontend = frontendFs
	// parse cobra flags
	rootCmd.ParseFlags(args)

	// unmarshal viper into config.Config
	config.Init()

	if config.Config().Debug != "" {
		debugo.SetDebug(config.Config().Debug)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
