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
	Tray               bool   `mapstructure:"tray"`
	Database           string `mapstructure:"database"`
	Debug              string `mapstructure:"debug"`
	Loglevel           string `mapstructure:"loglevel"`
	MaxConcurrentTasks uint   `mapstructure:"maxConcurrentTasks"`
	SendTelemetry      bool   `mapstructure:"sendTelemetry"`

	AI string `mapstructure:"ai"`
}

var config ConfigDefinition

func Init() {
	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("failed to unmarshal config: %s\n", err)
		os.Exit(1)
	}

	if config.Debug == "" {
		config.Debug = os.Getenv("DEBUGO")
	}
}

func Config() ConfigDefinition {
	return config
}
