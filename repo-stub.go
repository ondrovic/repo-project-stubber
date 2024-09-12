package main

import (
	"github-project-template/cmd/cli"
	sCli "github.com/ondrovic/common/utils/cli"
	"runtime"
)

func main() {
	if err := sCli.ClearTerminalScreen(runtime.GOOS); err != nil {
		return
	}

	if err := cli.RootCmd.Execute(); err != nil {
		return
	}
}
