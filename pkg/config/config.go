package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type ConfigDefinition struct {
	AppName    string `mapstructure:"appName"`
	AppVersion string `mapstructure:"appVersion"`

	FFMpeg string `mapstructure:"ffmpeg"`

	Port               uint   `mapstructure:"port"`
	Database           string `mapstructure:"database"`
	Debug              string `mapstructure:"debug"`
	MaxConcurrentTasks uint   `mapstructure:"maxConcurrentTasks"`
	SendTelemetry      bool   `mapstructure:"sendTelemetry"`
}

var config ConfigDefinition

func Init() {
	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("failed to unmarshal config: %s\n", err)
		os.Exit(1)
	}
}

func Config() ConfigDefinition {
	return config
}
