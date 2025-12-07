package main

import (
	_ "embed"

	"github.com/yaroslav-koval/hange/cmd"
)

//go:embed config.yaml
var buildConfig []byte

func main() {
	cmd.SetBuildConfig(buildConfig)

	cmd.Execute()
}
