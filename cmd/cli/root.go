package cli

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the root command of the CLI tool.
// It defines the command name, description, and usage for the tool that downloads GitHub repository contents for new projects.
var RootCmd = &cobra.Command{
	Use:   "repo-stub",
	Short: "A CLI tool to download GitHub repository contents when creating a new project.",
}

// Execute runs the root command and handles any errors that occur during execution.
// Returns: An error if the command execution fails.
func Execute() error {

	if err := RootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
