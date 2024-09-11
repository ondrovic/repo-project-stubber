package cli

import (
	"go.szostok.io/version/extension"
)

const (
	RepoOwner string = "ondrovic"
	RepoName  string = "repo-project-stubber"
)

func init() {
	RootCmd.AddCommand(
		extension.NewVersionCobraCmd(
			extension.WithUpgradeNotice(RepoOwner, RepoName),
		),
	)
}
