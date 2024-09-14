package cli

import (
	"go.szostok.io/version/extension"
)

const (
	// RepoOwner is the default GitHub repository owner for the CLI tool.
	RepoOwner string = "ondrovic"
	// RepoName is the default GitHub repository name for the CLI tool.
	RepoName string = "repo-project-stubber"
)

// init adds a new version command to the root command, including an upgrade notice.
// It initializes the version command with details about the repository owner and name.
// Parameters: None.
func init() {
	RootCmd.AddCommand(
		extension.NewVersionCobraCmd(
			extension.WithUpgradeNotice(RepoOwner, RepoName),
		),
	)
}
