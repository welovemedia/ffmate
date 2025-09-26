package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yosev/debugo"
)

var rootCmd = &cobra.Command{
	Use:   "ffmate",
	Short: "ffmate is a wrapper for ffmpeg",
}

func init() {
	rootCmd.PersistentFlags().StringP("debug", "d", "info:?,warn:?,error:?", "set debugo namespace (eg. '*')")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// setup debugo timestamp format
	debugo.SetTimestamp(&debugo.Timestamp{
		Format: "15:04:05.000",
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
