package cli

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "repo-stub",
	Short: "A CLI tool to download GitHub repository contents when creating a new project.",
}

func Execute() error {

	if err := RootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
