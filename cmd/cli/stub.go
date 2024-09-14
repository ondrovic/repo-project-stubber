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

// options holds the command-line flags for the CLI tool, represented by types.CliFlags.
var (
	// options is an instance of types.CliFlags used to manage command-line options.
	options = types.CliFlags{}
	// stubCmd represents a subcommand for the CLI tool.
	stubCmd = &cobra.Command{}
	// baseUrl is a string variable used to store the base URL for GitHub repository access.
	baseUrl string
)

// init initializes the `stubCmd` subcommand for the CLI tool.
// It sets up the command's usage, description, arguments, and execution function.
// It also binds the command-line flags to Viper for configuration management and adds the subcommand to the root command.
// Additionally, it constructs the base URL for GitHub repository access using the options provided.
// Parameters: None.
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

// initFlags sets up the flags for the given Cobra command, defining various options for repository access and project configuration.
// It adds flags for repository name, owner, branch name, GitHub token, project language, license type, and other options related to file inclusion and overwriting.
// Parameters:
// - cmd: A pointer to the Cobra command for which the flags are being defined.
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

// run is the execution function for the `stubCmd` subcommand.
// It sets the output directory from the command arguments, updates the base URL if a branch name other than "master" is specified,
// creates the output directory if it doesn't exist, and processes the repository based on the specified options.
// Parameters:
// - cmd: A pointer to the Cobra command being executed.
// - args: A slice of arguments provided to the command.
// Returns: An error if any issues occur during execution.
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
