package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print ffmate version",
	Run:   version,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func version(_ *cobra.Command, _ []string) {
	fmt.Printf("version: %s\n", viper.GetString("app.version"))
}
