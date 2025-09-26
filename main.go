package main

import (
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/v2/cmd"

	_ "embed"
)

//go:embed .version
var version string

// @title ffmate API
// @version
// @description	A wrapper around ffmpeg

// @contact.name We love media
// @contact.email sev@welovemedia.io

// @license.name AGPL-3.0
// @license.url https://opensource.org/license/agpl-v3

// @host localhost
// @BasePath /api/v1
func main() {
	viper.Set("app.name", "ffmate")
	viper.Set("app.version", version)

	cmd.Execute()
}
