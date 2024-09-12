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

func getRepoContents(url, path, token string) ([]types.GitHubItem, error) {

	if httpclient.Client == nil {
		httpclient.InitClient(token)
	}

	// Append the path to the URL if specified
	if path != "" {
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

func handleFileTypeContent(item types.GitHubItem, outputPath string, overwrite bool) error {
	switch strings.ToLower(item.Name) {
	case consts.README, consts.LICENSE, consts.GIT_IGNORE, consts.GIT_KEEP, consts.TODO:
		return nil
	default:
		// return nil
		return utils.SaveFile(item.DownloadURL, filepath.Join(outputPath, item.Path), overwrite)
	}
}

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
		return handleVersionFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.VSCODE_FILES:
		return handleVSCodeFiles(url, opts.OutputDirectory, opts.OverwriteFiles)
	case consts.WORKFLOW_FLIES:
		return handleWorkflowFiles(url, opts.ProjectLanguage, opts.OutputDirectory, opts.GithubToken, opts.OverwriteFiles)
	default:
		// TODO: fix the duplicates
		return ProcessRepository(url, item.Path, opts)
	}
}

func handleIgnoreFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	uri := fmt.Sprintf("%s/%s/%s/%s", url, consts.IGNORE_FILES, projectLanguage, consts.GIT_IGNORE)
	return utils.FileDownloader(uri, filepath.Join(outputPath, consts.GIT_IGNORE), overwrite)
}

func handleLicenseFiles(url, licenseType, outputPath string, overwrite bool) error {
	url = fmt.Sprintf("%s/%s/%s/%s", url, consts.LICENSE_FILES, licenseType, consts.LICENSE)
	return utils.FileDownloader(url, filepath.Join(outputPath, consts.LICENSE), overwrite)
}

func handleMakeFiles(url, projectLanguage, outputPath string, overwrite, includeMakefile bool) error {
	if !includeMakefile {
		return nil
	}

	url = fmt.Sprintf("%s/%s/%s/%s", url, consts.MAKE_FILES, projectLanguage, consts.MAKEFILE)

	return utils.FileDownloader(url, filepath.Join(outputPath, consts.MAKEFILE), overwrite)
}

func handleReadmeFiles(url, licenseType, outputPath string, overwrite bool) error {
	url = fmt.Sprintf("%s/%s/%s/%s", url, consts.README_FILES, licenseType, consts.README)
	return utils.FileDownloader(url, filepath.Join(outputPath, consts.README), overwrite)
}

func handleTodoFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	url = fmt.Sprintf("%s/%s/%s/%s", url, consts.TODO_FILES, projectLanguage, consts.TODO)
	return utils.FileDownloader(url, filepath.Join(outputPath, consts.TODO), overwrite)
}

func handleVSCodeFiles(url, outputPath string, overwrite bool) error {
	url = fmt.Sprintf("%s/%s/commands.json", url, consts.VSCODE_FILES)
	return utils.FileDownloader(url, filepath.Join(outputPath, consts.VSCODE, "commands.json"), overwrite)
}

func handleVersionFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	if versionFile, err := utils.GetVersionFile(projectLanguage); err != nil {
		return err
	} else if versionFile == "" {
		return fmt.Errorf("no version file for %s", projectLanguage)
	} else {
		url = fmt.Sprintf("%s/%s/%s/%s", url, consts.VERSION_FILES, projectLanguage, versionFile)
		return utils.FileDownloader(url, filepath.Join(outputPath, versionFile), overwrite)
	}
}

func handleReleaseFiles(url, projectLanguage, outputPath string, overwrite bool) error {
	if releaseFile, err := utils.GetReleaseFile(projectLanguage); err != nil {
		return err
	} else if releaseFile == "" {
		return fmt.Errorf("no release file for %s", projectLanguage)
	} else {
		url = fmt.Sprintf("%s/%s/%s/%s", url, consts.RELEASE_FILES, projectLanguage, releaseFile)
		return utils.FileDownloader(url, filepath.Join(outputPath, fmt.Sprintf(".%s", releaseFile)), overwrite)
	}
}

func handleWorkflowFiles(url, projectLanguage, outputPath, token string, overwrite bool) error {
	url = fmt.Sprintf("%s/%s/%s", url, consts.WORKFLOW_FLIES, projectLanguage)

	contents, err := getRepoContents(url, "", token)
	if err != nil {
		return err
	}

	for _, item := range contents {
		if item.Type == consts.FILE_TYPE && strings.HasSuffix(item.Name, consts.YML) {
			outputPath = filepath.Join(outputPath, consts.GIT_HUB, consts.WORKFLOW, item.Name)
			if err := utils.SaveFile(item.DownloadURL, outputPath, overwrite); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: move all if <items> == "" to validation return type func
