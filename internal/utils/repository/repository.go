package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gookit/color"

	"github-project-template/internal/consts"
	"github-project-template/internal/httpclient"
	"github-project-template/internal/types"
	"github-project-template/internal/utils"
)

var (
	wg sync.WaitGroup
)

// getRepoContents: Retrieves the contents of a GitHub repository based on the provided URL and path, using a GitHub token for authentication.
func getRepoContents(url, path, token string) ([]types.GitHubItem, error) {

	if httpclient.Client == nil {
		httpclient.InitClient(token)
	}

	// Append the path to the URL if specified
	if path != consts.EMPTY_STRING {
		url = fmt.Sprintf("%s/%s", url, path)
	}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("failed to create request %s: %v", url, err)
	}

	// Perform the request using the global client
	resp, err := httpclient.Client.Do(req)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("failed to get contents from '%s': %v", url, err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(": %s", utils.SetColor(color.FgLightRed, resp.Status))
		// return nil, fmt.Errorf("failed to get contents %s: HTTP status %d", url, resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// return nil, fmt.Errorf("failed to read response body: %v", err)
		return nil, err
	}

	// Try to unmarshal the body as an array
	var contents []types.GitHubItem
	err = json.Unmarshal(body, &contents)
	if err == nil {
		// Return the contents, whether it's a single item or multiple items
		return contents, nil
	}

	// If it's not an array, try to unmarshal as a single object
	var singleItem types.GitHubItem
	err = json.Unmarshal(body, &singleItem)
	if err != nil {
		// return nil, fmt.Errorf("failed to decode response %s: %v", url, err)
		return nil, err
	}

	return []types.GitHubItem{singleItem}, nil
}

// ProcessRepository: Processes the contents of a repository by handling files and directories according to their types and options passed from the CLI.
func ProcessRepository(url, path string, opts types.CliFlags) error {
	contents, err := getRepoContents(url, path, opts.GithubToken)
	if err != nil {
		return fmt.Errorf("%s %s %v", utils.SetColor(color.FgLightRed, "getting repo contents"), utils.SetColor(color.FgLightCyan, url), err)
		// return err
	}

	if len(contents) == 0 {
		fmt.Println("no contents found")
		return nil
	}

	for _, item := range contents {
		switch item.Type {
		case consts.FILE_TYPE:
			wg.Add(1)
			go func(item types.GitHubItem) {
				defer wg.Done()
				if err := handleFileTypeContent(item, opts.OutputDirectory, opts.OverwriteFiles); err != nil {
					// fmt.Printf("error handling %s file %s: %v\n", url, item.Path, err)
					// return err
					fmt.Println(err)
				}
			}(item)
		case consts.DIR_TYPE:
			if err := handleDirectoryTypeContent(url, opts, item); err != nil {
				// fmt.Printf("error handling %s directory %s: %v\n", url, item.Path, err)
				fmt.Println(err)
			}
		default:
			return fmt.Errorf("unknown item.Type: %s found at %s", item.Type, item.Path)
		}
	}
	wg.Wait()
	return nil
}

// handleFileTypeContent: Handles the processing of file-type contents by saving files based on the item’s name and provided options.
func handleFileTypeContent(item types.GitHubItem, outputPath string, overwrite bool) error {
	switch item.Name {
	case consts.README, consts.LICENSE, consts.GIT_IGNORE, consts.GIT_KEEP, consts.TODO:
		// skip these they get handled in their relative funcs below
		return nil
	default:
		return utils.SaveFile(item.DownloadURL, filepath.Join(outputPath, item.Path), overwrite)
	}
}

// handleDirectoryTypeContent: Manages the processing of directory-type contents, calling the appropriate handler functions for specific directory types.
func handleDirectoryTypeContent(url string, opts types.CliFlags, item types.GitHubItem) error {
	switch item.Name {
	case consts.IGNORE_FILES:
		return handleIgnoreFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.LICENSE_FILES:
		return handleLicenseFiles(url, opts.LicenseType, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.MAKE_FILES:
		return handleMakeFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.OverwriteFiles, opts.IncludeMakefile)
	case consts.README_FILES:
		return handleReadmeFiles(url, opts.LicenseType, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.TODO_FILES:
		return handleTodoFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.RELEASE_FILES:
		return handleReleaseFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.VERSION_FILES:
		return handleVersionFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.OverwriteFiles, opts.IncludeVersionFile)
	case consts.VSCODE_FILES:
		return handleVSCodeFiles(url, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.WORKFLOW_FLIES:
		return handleWorkflowFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.GithubToken, opts.OverwriteFiles)
	default:
		return ProcessRepository(url, item.Path, opts)
	}
}

// handleIgnoreFiles: Processes and saves .gitignore files based on the project language.
func handleIgnoreFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.IGNORE_FILES, projectLanguage, consts.GIT_IGNORE)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.GIT_IGNORE), overwrite)
}

// handleLicenseFiles: Processes and saves license files based on the license type.
func handleLicenseFiles(url, licenseType, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.LICENSE_FILES, licenseType, consts.LICENSE)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.LICENSE), overwrite)
}

// handleMakeFiles: Processes and saves Makefiles if the includeMakefile option is enabled.
func handleMakeFiles(url, projectLanguage, outputPath string, overwrite, includeMakefile bool) error {
	if !includeMakefile {
		return nil
	}
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.MAKE_FILES, projectLanguage, consts.MAKEFILE)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.MAKEFILE), overwrite)
}

// handleReadmeFiles: Processes and saves README files based on the license type.
func handleReadmeFiles(url, licenseType, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.README_FILES, licenseType, consts.README)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.README), overwrite)
}

// handleTodoFiles: Processes and saves TODO files based on the project language.
func handleTodoFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.TODO_FILES, projectLanguage, consts.TODO)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.TODO), overwrite)
}

// handleVSCodeFiles: Processes and saves VSCode configuration files (like commands.json).
func handleVSCodeFiles(url, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/commands.json", url, consts.VSCODE_FILES)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.VSCODE, "commands.json"), overwrite)
}

// handleVersionFiles: Processes and saves version files if the includeVersionFile option is enabled, based on the project language.
func handleVersionFiles(url, projectLanguage, outputPath string, overwrite, includeVersionFile bool) error {
	if !includeVersionFile {
		return nil
	}
	versionFile, err := utils.GetVersionFile(projectLanguage)
	if err != nil {
		return err
	}
	if versionFile == consts.EMPTY_STRING {
		return fmt.Errorf("no version file for %s", projectLanguage)
	}

	// Construct the content URL
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.VERSION_FILES, projectLanguage, versionFile)

	// Get the download URL
	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	// Save the file
	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, versionFile), overwrite)
}

// handleReleaseFiles: Processes and saves release files for the specified project language.
func handleReleaseFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	// Get the release file for the specified language
	releaseFile, err := utils.GetReleaseFile(projectLanguage)
	if err != nil {
		return err
	}
	if releaseFile == consts.EMPTY_STRING {
		return fmt.Errorf("no release file for %s", projectLanguage)
	}

	// Construct the content URL
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.RELEASE_FILES, projectLanguage, releaseFile)

	// Get the download URL
	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	// Save the file
	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, fmt.Sprintf(".%s", releaseFile)), overwrite)
}

// handleWorkflowFiles: Processes and saves GitHub workflow YAML files for the project language in the appropriate directory.
func handleWorkflowFiles(url, projectLanguage, outputPath, token string, overwrite bool) error {
	// Build the initial URL for workflow files
	url = fmt.Sprintf("%s/%s/%s", url, consts.WORKFLOW_FLIES, projectLanguage)

	// Fetch the contents of the repository
	contents, err := getRepoContents(url, consts.EMPTY_STRING, token)
	if err != nil {
		return err
	}

	// Iterate through the contents
	for _, item := range contents {
		if item.Type == consts.FILE_TYPE && strings.HasSuffix(item.Name, consts.YML) {
			// Join the output path correctly for each file
			fileOutputPath := filepath.Join(outputPath, consts.GIT_HUB, consts.WORKFLOW, item.Name)

			// Save the file
			if err := utils.SaveFile(item.DownloadURL, fileOutputPath, overwrite); err != nil {
				return err
			}
		}
	}

	return nil
}
