package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ondrovic/common/utils/cli"
)

type GitHubItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

var (
	repoOwner              string
	repoName               string
	branch                 string
	outputDirectory        string
	includeMakeFile        bool
	projectLanguage        string
	overwriteExistingFiles bool
	licenseType            string
	githubToken            string
	multi                  pterm.MultiPrinter
	wg                     sync.WaitGroup
	client                 *http.Client
)

func main() {
	if err := cli.ClearTerminalScreen(runtime.GOOS); err != nil {
		pterm.Error.Print(err)
		return
	}

	rootCmd := &cobra.Command{
		Use:   "repo-stub",
		Short: "A CLI tool to download GitHub repository contents when creating a new project.",
		Run:   run,
	}

	initFlags(rootCmd)
	viper.BindPFlags(rootCmd.Flags())

	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Print(err)
	}
}

func initFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&repoOwner, "owner", "o", "ondrovic", "Repository owner")
	cmd.Flags().StringVarP(&repoName, "repo", "r", "vscode", "Repository name")
	cmd.Flags().StringVarP(&branch, "branch", "b", "master", "Branch name")
	cmd.Flags().StringVarP(&outputDirectory, "output", "d", "", "Output directory")
	cmd.Flags().BoolVarP(&includeMakeFile, "makefile", "m", false, "Does your project need a makefile?")
	cmd.Flags().StringVarP(&projectLanguage, "project-language", "p", "go", "What language is your app?")
	cmd.Flags().BoolVarP(&overwriteExistingFiles, "overwrite", "w", false, "Overwrite existing files?")
	cmd.Flags().StringVarP(&licenseType, "license", "l", "mit", "What license do you want to use?")
	cmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub API token")
	cmd.MarkFlagRequired("output")
}

func run(cmd *cobra.Command, args []string) {
	client = &http.Client{}

	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", repoOwner, repoName)
	if branch != "master" {
		baseURL += "?ref=" + branch
	}

	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		pterm.Error.Printf("Error creating output directory: %v\n", err)
		return
	}

	multi = pterm.DefaultMultiPrinter
	multi.Start()
	defer multi.Stop()

	if err := processRepository(baseURL, ""); err != nil {
		pterm.Error.Printf("Error processing repository %s: %v\n", baseURL, err)
	}
}

func getRepoContents(repoUri, path string) ([]GitHubItem, error) {
	url := repoUri
	if path != "" {
		url = fmt.Sprintf("%s/%s", repoUri, path)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request %s: %v", url, err)
	}

	if githubToken != "" {
		req.Header.Set("Authorization", "token "+githubToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents from '%s': %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get contents %s: HTTP status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var contents []GitHubItem

	// Try to unmarshal as an array first
	err = json.Unmarshal(body, &contents)
	if err == nil {
		// If it's an array with a single item, return that item
		if len(contents) == 1 {
			return contents, nil
		}
		// If it's an array with multiple items, return all items
		return contents, nil
	}

	// If it's not an array, try to unmarshal as a single object
	var singleItem GitHubItem
	err = json.Unmarshal(body, &singleItem)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response %s: %v", url, err)
	}

	// Return the single item as an array with one element
	return []GitHubItem{singleItem}, nil
}

func processRepository(baseURL, path string) error {
	contents, err := getRepoContents(baseURL, path)
	if err != nil {
		return fmt.Errorf("error getting GitHub contents %s: %v", baseURL, err)
	}

	if len(contents) == 0 {
		pterm.Warning.Printf("No contents found in the repository path: %s\n", path)
		return nil
	}

	for _, item := range contents {
		switch item.Type {
		case "file":
			wg.Add(1)
			go func(item GitHubItem) {
				defer wg.Done()
				if err := handleFileTypeContent(item); err != nil {
					pterm.Error.Printf("Error handling %s file %s: %v\n", baseURL, item.Path, err)
				}
			}(item)
		case "dir":
			if err := handleDirectoryTypeContent(baseURL, item); err != nil {
				pterm.Error.Printf("Error handling %s directory %s: %v\n", baseURL, item.Path, err)
			}
		}
	}

	wg.Wait()
	return nil
}

func handleFileTypeContent(item GitHubItem) error {
	switch strings.ToLower(item.Name) {
	case "readme.md", "license", ".gitignore", "todo":
		// These files are handled separately
		return nil
	case ".gitkeep":
		// Skip .gitkeep files
		return nil
	default:
		return downloadAndSaveFile(item.DownloadURL, filepath.Join(outputDirectory, item.Path), overwriteExistingFiles)
	}
}

func handleDirectoryTypeContent(baseURL string, item GitHubItem) error {
	switch strings.ToLower(item.Name) {
	case ".ignorefiles":
		return handleIgnoreFile(baseURL)
	case ".licensefiles":
		return handleLicenseFile(baseURL)
	case ".makefiles":
		return handleMakeFiles(baseURL)
	case ".readmefiles":
		return handleReadmeFile(baseURL)
	case ".todo":
		return handleTodoFile(baseURL)
	case ".vscode":
		return handleVSCodeFiles(baseURL)
	case ".workflows":
		return handleWorkflowFiles(baseURL)
	case ".releasers":
		return handleReleaserFiles(baseURL)
	default:
		return processRepository(baseURL, item.Path)
	}
}

func handleIgnoreFile(baseURL string) error {
	ignoreApiURL := fmt.Sprintf("%s/.ignorefiles", baseURL)
	item, err := getGitHubFileInfo(ignoreApiURL)
	if err != nil {
		return fmt.Errorf("failed to get .gitignore file info: %v", err)
	}

	return downloadAndSaveFile(item.DownloadURL, filepath.Join(outputDirectory, ".gitignore"), overwriteExistingFiles)
}

func handleLicenseFile(baseURL string) error {
	if licenseType == "" {
		return nil
	}
	licenseApiURL := fmt.Sprintf("%s/.licensefiles/%s/LICENSE", baseURL, licenseType)
	return downloadFileFromGitHub(licenseApiURL, filepath.Join(outputDirectory, "LICENSE"), overwriteExistingFiles)
}

func handleMakeFiles(baseURL string) error {
	if !includeMakeFile {
		return nil
	}
	makefileApiURL := fmt.Sprintf("%s/.makefiles/%s/Makefile", baseURL, projectLanguage)
	return downloadFileFromGitHub(makefileApiURL, filepath.Join(outputDirectory, "Makefile"), overwriteExistingFiles)
}

func handleReadmeFile(baseURL string) error {
	readmeApiURL := fmt.Sprintf("%s/.readmefiles/%s/README.md", baseURL, licenseType)
	return downloadFileFromGitHub(readmeApiURL, filepath.Join(outputDirectory, "README.md"), overwriteExistingFiles)
}

func handleTodoFile(baseURL string) error {
	todoApiURL := fmt.Sprintf("%s/.todo/TODO", baseURL)
	return downloadFileFromGitHub(todoApiURL, filepath.Join(outputDirectory, "TODO"), overwriteExistingFiles)
}

func handleVSCodeFiles(baseURL string) error {
	vscodeApiURL := fmt.Sprintf("%s/.vscode/commands.json", baseURL)
	return downloadFileFromGitHub(vscodeApiURL, filepath.Join(outputDirectory, ".vscode", "commands.json"), overwriteExistingFiles)
}

func handleReleaserFiles(baseURL string) error {
	releaserApiURL := fmt.Sprintf("%s/.releasers/%s/goreleaser.yaml", baseURL, projectLanguage)
	return downloadFileFromGitHub(releaserApiURL, filepath.Join(outputDirectory, "goreleaser.yaml"), overwriteExistingFiles)
}

func handleWorkflowFiles(baseURL string) error {
	workflowsApiURL := fmt.Sprintf("%s/.workflows/%s", baseURL, projectLanguage)
	workflowContents, err := getRepoContents(workflowsApiURL, "")
	if err != nil {
		return err
	}

	for _, item := range workflowContents {
		if item.Type == "file" && strings.HasSuffix(item.Name, ".yml") {
			destPath := filepath.Join(outputDirectory, ".github", "workflows", item.Name)
			if err := downloadAndSaveFile(item.DownloadURL, destPath, overwriteExistingFiles); err != nil {
				pterm.Error.Printf("Error downloading workflow file %s: %v\n", item.Name, err)
			}
		}
	}

	return nil
}

func downloadFileFromGitHub(apiURL, destPath string, overwrite bool) error {
	item, err := getGitHubFileInfo(apiURL)
	if err != nil {
		return fmt.Errorf("failed to get file info %s: %v", apiURL, err)
	}

	return downloadAndSaveFile(item.DownloadURL, destPath, overwrite)
}

func getGitHubFileInfo(apiURL string) (*GitHubItem, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info from '%s': %v", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get file info: HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var items []GitHubItem
	err = json.Unmarshal(body, &items)
	if err == nil && len(items) > 0 {
		// If it's an array, return the first item
		return &items[0], nil
	}

	// If it's not an array, try to unmarshal as a single item
	var item GitHubItem
	err = json.Unmarshal(body, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file info: %v", err)
	}

	return &item, nil
}

func downloadAndSaveFile(url, destPath string, overwrite bool) error {
	if _, err := os.Stat(destPath); err == nil && !overwrite {
		pterm.Info.Printf("Skipping '%s'\n", destPath)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory structure for '%s': %v", destPath, err)
	}

	spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start(fmt.Sprintf("Downloading '%s'", filepath.Base(destPath)))

	resp, err := http.Get(url)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to download '%s': %v", filepath.Base(destPath), err))
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		spinner.Fail(fmt.Sprintf("Failed to download '%s': HTTP status %d", filepath.Base(destPath), resp.StatusCode))
		return fmt.Errorf("failed to download file: HTTP status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to create file '%s': %v", destPath, err))
		return fmt.Errorf("failed to create file '%s': %v", destPath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to write file '%s': %v", destPath, err))
		return fmt.Errorf("failed to write file '%s': %v", destPath, err)
	}

	spinner.Success(fmt.Sprintf("Downloaded '%s' to '%s'", filepath.Base(destPath), destPath))
	return nil
}
