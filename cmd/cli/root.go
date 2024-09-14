package cli

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "repo-stub",
	Short: "A CLI tool to download GitHub repository contents when creating a new project.",
}

// Execute runs the RootCmd command of the CLI. It attempts to execute the root command
// and returns an error if one occurs during execution. If the command runs successfully,
// it returns nil.
func Execute() error {

	if err := RootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
