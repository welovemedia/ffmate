package main

import (
	"os"

	_ "embed"

	"github.com/welovemedia/ffmate/cmd"
)

//go:embed .version
var version string

func main() {
	cmd.Execute(os.Args, version)
}
