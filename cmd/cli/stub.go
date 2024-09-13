package cli

import (
	"fmt"
	"github-project-template/internal/consts"
	"github-project-template/internal/types"
	"github-project-template/internal/utils/repository"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	options = types.CliFlags{}
	stubCmd = &cobra.Command{}
	baseUrl string
)

func init() {
	stubCmd = &cobra.Command{
		Use:   "stub <output-directory> [flags]",
		Short: "Stub project",
		Long:  "Stub project with templated files based on options",
		Args:  cobra.ExactArgs(1),
		RunE:  run,
	}

	initFlags(stubCmd)
	viper.BindPFlags(stubCmd.Flags())
	RootCmd.AddCommand(stubCmd)

	baseUrl = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", options.RepoOwner, options.RepoName)
}

func initFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&options.RepoName, "repo-name", "r", "vscode", "Name of the repository")
	cmd.Flags().StringVarP(&options.RepoOwner, "repo-owner", "o", "ondrovic", "Owner of the repository")
	cmd.Flags().StringVarP(&options.BranchName, "branch-name", "b", "master", "Branch name you wish to pull from")
	cmd.Flags().StringVarP(&options.GithubToken, "github-token", "t", consts.EMPTY_STRING, "Github API token")
	cmd.Flags().StringVarP(&options.ProjectLanguage, "project-language", "p", "go", "What language is your app in")
	cmd.Flags().StringVarP(&options.LicenseType, "license-type", "l", "mit", "What license are you using")
	cmd.Flags().BoolVarP(&options.IncludeMakefile, "include-makefile", "m", false, "Include a Makefile")
	cmd.Flags().BoolVarP(&options.IncludeVersionFile, "include-version-file", "v", false, "Include a version file")
	cmd.Flags().BoolVarP(&options.OverwriteFiles, "overwrite-files", "w", false, "Overwrite existing files")
}

func run(cmd *cobra.Command, args []string) error {

	options.OutputDirectory = args[0]

	if options.BranchName != "master" {
		baseUrl = fmt.Sprintf("%s?ref=%s", baseUrl, options.BranchName)
	}

	if err := os.MkdirAll(options.OutputDirectory, 0755); err != nil {
		fmt.Println(err)
	}

	if err := repository.ProcessRepository(baseUrl, consts.EMPTY_STRING, options); err != nil {
		fmt.Println(err)
	}

	return nil
}
