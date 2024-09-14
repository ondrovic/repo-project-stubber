package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	// "github.com/gookit/color"

	"github-project-template/internal/consts"
	"github-project-template/internal/httpclient"
	"github-project-template/internal/types"
	"github-project-template/internal/utils"
)

var (
	wg sync.WaitGroup
)

// getRepoContents: Retrieves the contents of a GitHub repository based on the provided URL and path, using a GitHub token for authentication.
// func getRepoContents(url, path, token string) ([]types.GitHubItem, error) {

// 	if httpclient.Client == nil {
// 		httpclient.InitClient(token)
// 	}

// 	// Append the path to the URL if specified
// 	if path != consts.EMPTY_STRING {
// 		url = fmt.Sprintf("%s/%s", url, path)
// 	}

// 	// Create a new request
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 		// return nil, fmt.Errorf("failed to create request %s: %v", url, err)
// 	}

// 	// Perform the request using the global client
// 	resp, err := httpclient.Client.Do(req)
// 	if err != nil {
// 		return nil, err
// 		// return nil, fmt.Errorf("failed to get contents from '%s': %v", url, err)
// 	}
// 	defer resp.Body.Close()

// 	// Check the response status code
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf(": %s", utils.SetColor(color.FgLightRed, resp.Status))
// 		// return nil, fmt.Errorf("failed to get contents %s: HTTP status %d", url, resp.StatusCode)
// 	}

// 	// Read the response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		// return nil, fmt.Errorf("failed to read response body: %v", err)
// 		return nil, err
// 	}

// 	// Try to unmarshal the body as an array
// 	var contents []types.GitHubItem
// 	err = json.Unmarshal(body, &contents)
// 	if err == nil {
// 		// Return the contents, whether it's a single item or multiple items
// 		return contents, nil
// 	}

// 	// If it's not an array, try to unmarshal as a single object
// 	var singleItem types.GitHubItem
// 	err = json.Unmarshal(body, &singleItem)
// 	if err != nil {
// 		// return nil, fmt.Errorf("failed to decode response %s: %v", url, err)
// 		return nil, err
// 	}

// 	return []types.GitHubItem{singleItem}, nil
// }

func getRepoContents(url, path, token string) ([]types.GitHubItem, error) {
	httpClient, err := httpclient.InitClient(token)
	if err != nil {
		return nil, err
	}

	url = appendPathToUrl(url, path)

	req, err := createRequest(url, "GET")
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed with %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var contents []types.GitHubItem
	err = json.Unmarshal(body, &contents)
	if err != nil {
		var singleItem types.GitHubItem
		err = json.Unmarshal(body, &singleItem)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response %s: %v", url, err)
		}
		return []types.GitHubItem{singleItem}, nil
	}

	return contents, nil
}

// appendPathToUrl appends the given path to the URL if the path is not an empty string.
// If the path is not empty, it ensures the URL is properly formatted with a "/" between the base URL and the path.
// Parameters:
// - url: The base URL as a string.
// - path: The path to append to the URL as a string.
// Returns: A string representing the full URL with the appended path.
func appendPathToUrl(url, path string) string {
	if path != consts.EMPTY_STRING {
		url = fmt.Sprintf("%s/%s", url, path)
	}
	return url
}

// createRequest creates a new HTTP request with the given URL and method.
// It returns the request object and any error encountered during its creation.
// Parameters:
// - url: The URL for the request as a string.
// - method: The HTTP method to use (e.g., "GET", "POST").
// Returns: A pointer to an http.Request and an error (if any).
func createRequest(url, method string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// ProcessRepository processes the contents of a repository at the given URL and path, based on the provided CLI flags.
// It retrieves the contents of the repository and processes each item based on its type (file or directory).
// Files are handled concurrently, while directories are processed sequentially.
// Parameters:
// - url: The repository URL as a string.
// - path: The path within the repository as a string.
// - opts: CLI options of type types.CliFlags, including settings like GitHub token, output directory, and overwrite flag.
// Returns: An error if any issues occur during processing or repository content retrieval.
func ProcessRepository(url, path string, opts types.CliFlags) error {
	contents, err := getRepoContents(url, path, opts.GithubToken)
	if err != nil {
		return err
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
					fmt.Println(err)
				}
			}(item)
		case consts.DIR_TYPE:
			if err := handleDirectoryTypeContent(url, opts, item); err != nil {
				fmt.Println(err)
			}
		default:
			return fmt.Errorf("unknown item.Type: %s found at %s", item.Type, item.Path)
		}
	}
	wg.Wait()
	return nil
}

// handleFileTypeContent processes a GitHub item of type "file".
// It skips certain special files (e.g., README, LICENSE) and saves other files to the specified output path.
// Parameters:
// - item: The GitHub item to process, of type types.GitHubItem.
// - outputPath: The directory where the file should be saved.
// - overwrite: A boolean indicating whether existing files should be overwritten.
// Returns: An error if any issues occur during file saving.
func handleFileTypeContent(item types.GitHubItem, outputPath string, overwrite bool) error {
	switch item.Name {
	case consts.README, consts.LICENSE, consts.GIT_IGNORE, consts.GIT_KEEP, consts.TODO:
		// skip these they get handled in their relative funcs below
		return nil
	default:
		return utils.SaveFile(item.DownloadURL, filepath.Join(outputPath, item.Path), overwrite)
	}
}

// handleDirectoryTypeContent processes a GitHub item of type "directory".
// Depending on the directory name, it delegates handling to specific functions for various types of files (e.g., ignore files, license files).
// If the directory doesn't match any known special cases, it processes the repository recursively.
// Parameters:
// - url: The repository URL as a string.
// - opts: CLI options of type types.CliFlags, including settings like project language, output directory, and overwrite flag.
// - item: The GitHub item to process, of type types.GitHubItem.
// Returns: An error if any issues occur during directory processing or file handling.
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

// handleIgnoreFiles processes and saves the ignore files (e.g., .gitignore) for the specified project language.
// It constructs the download URL based on the project language and fetches the file from the repository.
// Parameters:
// - url: The base repository URL as a string.
// - projectLanguage: The programming language of the project, used to determine the ignore file to fetch.
// - outputPath: The directory where the ignore file should be saved.
// - overwrite: A boolean indicating whether to overwrite an existing ignore file.
// Returns: An error if any issues occur during the file download or saving process.
func handleIgnoreFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.IGNORE_FILES, projectLanguage, consts.GIT_IGNORE)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.GIT_IGNORE), overwrite)
}

// handleLicenseFiles processes and saves the license file for the specified license type.
// It constructs the download URL based on the license type and fetches the license file from the repository.
// Parameters:
// - url: The base repository URL as a string.
// - licenseType: The type of license to fetch (e.g., MIT, GPL).
// - outputPath: The directory where the license file should be saved.
// - overwrite: A boolean indicating whether to overwrite an existing license file.
// Returns: An error if any issues occur during the file download or saving process.
func handleLicenseFiles(url, licenseType, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.LICENSE_FILES, licenseType, consts.LICENSE)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.LICENSE), overwrite)
}

// handleMakeFiles processes and saves the Makefile for the specified project language, if the includeMakefile option is set to true.
// It constructs the download URL based on the project language and fetches the Makefile from the repository.
// If includeMakefile is false, the function returns without performing any actions.
// Parameters:
// - url: The base repository URL as a string.
// - projectLanguage: The programming language of the project, used to determine the appropriate Makefile to fetch.
// - outputPath: The directory where the Makefile should be saved.
// - overwrite: A boolean indicating whether to overwrite an existing Makefile.
// - includeMakefile: A boolean indicating whether to include the Makefile in the process.
// Returns: An error if any issues occur during the file download or saving process.
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

// handleReadmeFiles processes and saves the README file for the specified license type.
// It constructs the download URL based on the license type and fetches the README file from the repository.
// Parameters:
// - url: The base repository URL as a string.
// - licenseType: The type of license, used to determine the appropriate README file to fetch.
// - outputPath: The directory where the README file should be saved.
// - overwrite: A boolean indicating whether to overwrite an existing README file.
// Returns: An error if any issues occur during the file download or saving process.
func handleReadmeFiles(url, licenseType, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.README_FILES, licenseType, consts.README)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.README), overwrite)
}

// handleTodoFiles processes and saves the TODO file for the specified project language.
// It constructs the download URL based on the project language and fetches the TODO file from the repository.
// Parameters:
// - url: The base repository URL as a string.
// - projectLanguage: The programming language of the project, used to determine the appropriate TODO file to fetch.
// - outputPath: The directory where the TODO file should be saved.
// - overwrite: A boolean indicating whether to overwrite an existing TODO file.
// Returns: An error if any issues occur during the file download or saving process.
func handleTodoFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/%s/%s", url, consts.TODO_FILES, projectLanguage, consts.TODO)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.TODO), overwrite)
}

// handleVSCodeFiles processes and saves the VSCode configuration file (commands.json) for the repository.
// It constructs the download URL for the VSCode files and fetches the commands.json file from the repository.
// Parameters:
// - url: The base repository URL as a string.
// - outputPath: The directory where the VSCode commands.json file should be saved.
// - overwrite: A boolean indicating whether to overwrite an existing commands.json file.
// Returns: An error if any issues occur during the file download or saving process.
func handleVSCodeFiles(url, outputPath string, overwrite bool) error {
	contentUrl := fmt.Sprintf("%s/%s/commands.json", url, consts.VSCODE_FILES)

	downloadUrl, err := utils.GrabDownloadUrl(contentUrl)
	if err != nil {
		return err
	}

	return utils.SaveFile(downloadUrl, filepath.Join(outputPath, consts.VSCODE, "commands.json"), overwrite)
}

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

// handleVersionFiles processes and saves the version file for the specified project language, if the includeVersionFile option is set to true.
// It retrieves the appropriate version file name based on the project language and constructs the download URL to fetch the file from the repository.
// If includeVersionFile is false, the function returns without performing any actions.
// Parameters:
// - url: The base repository URL as a string.
// - projectLanguage: The programming language of the project, used to determine the appropriate version file to fetch.
// - outputPath: The directory where the version file should be saved.
// - overwrite: A boolean indicating whether to overwrite an existing version file.
// - includeVersionFile: A boolean indicating whether to include the version file in the process.
// Returns: An error if any issues occur during the file retrieval or saving process.
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

// handleWorkflowFiles processes and saves workflow files (with .yml extension) for the specified project language.
// It constructs the URL to fetch workflow files, retrieves the contents of the repository, and saves each workflow file to the specified output directory.
// Parameters:
// - url: The base repository URL as a string.
// - projectLanguage: The programming language of the project, used to determine the workflow files to fetch.
// - outputPath: The directory where the workflow files should be saved.
// - token: The authentication token to access private repositories.
// - overwrite: A boolean indicating whether to overwrite existing workflow files.
// Returns: An error if any issues occur during file retrieval or saving.
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
