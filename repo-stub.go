package main

import (
	"github-project-template/cmd/cli"
	sCli "github.com/ondrovic/common/utils/cli"
	"github.com/pterm/pterm"
	"runtime"
)

func main() {
	if err := sCli.ClearTerminalScreen(runtime.GOOS); err != nil {
		pterm.Error.Print(err)
		return
	}

	if err := cli.RootCmd.Execute(); err != nil {
		return
	}
}
