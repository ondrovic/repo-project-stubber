package main

import (
	"github-project-template/cmd/cli"
	"runtime"

	sCli "github.com/ondrovic/common/utils/cli"
)

// main is the entry point for the application. It clears the terminal screen based on the
// operating system and then executes the root command of the CLI. If any error occurs during
// these operations, the function returns without further action.
func main() {
	if err := sCli.ClearTerminalScreen(runtime.GOOS); err != nil {
		return
	}

	if err := cli.RootCmd.Execute(); err != nil {
		return
	}
}
