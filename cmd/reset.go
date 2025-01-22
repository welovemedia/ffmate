package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/sev"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset running jobs",
	Run:   reset,
}

func init() {
	resetCmd.PersistentFlags().StringP("database", "", "db.sqlite", "the path do the database")
	viper.BindPFlag("database", resetCmd.PersistentFlags().Lookup("database"))
	rootCmd.AddCommand(resetCmd)
}

func reset(cmd *cobra.Command, args []string) {
	s := sev.New("ffmate", config.Config().AppVersion, config.Config().Database, 0)
	s.DB().Model(&model.Task{}).
		Where("status = ?", "RUNNING").
		Updates(map[string]interface{}{
			"status":   "DONE_CANCELED",
			"progress": 100,
		})
	s.Logger().Infof("All running jobs have been reset")
}
