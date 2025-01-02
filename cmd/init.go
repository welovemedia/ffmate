package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var port uint
var dbPath string
var debug bool
var appVersion string
var concurrentTasks uint
var sendTelemetry bool

var rootCmd = &cobra.Command{
	Use:               "ffmate",
	Short:             "ffmate is a wrapper for ffmpeg",
	Long:              "ffmate is a wrapper for ffmmpeg that adds a queue system on top of it",
	DisableAutoGenTag: true,
	CompletionOptions: cobra.CompletionOptions{},
}

func Execute(args []string, version string) {
	appVersion = version
	rootCmd.PersistentFlags().StringVarP(&dbPath, "database", "", "db.sqlite", "the path do the database")
	rootCmd.PersistentFlags().UintVarP(&port, "port", "p", 3000, "set the port for the server to run")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", true, "emable debugging")
	rootCmd.PersistentFlags().UintVarP(&concurrentTasks, "max-concurrent-tasks", "m", 1, "define maximum concurrent running tasks")
	rootCmd.PersistentFlags().BoolVarP(&sendTelemetry, "send-telemetry", "", true, "enable sending anonymous telemtry data")

	rootCmd.ParseFlags(args)

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
