package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github-project-template/internal/consts"
	"github-project-template/internal/httpclient"
	"github-project-template/internal/spinner"
	"github-project-template/internal/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/color"
)

// SetColor: Formats the given item as a string and applies the specified color to the output.
func SetColor(col color.Color, item interface{}) string {
	return col.Sprintf("%v", item)
}

// GrabDownloadUrl: Retrieves the download URL from a given API URL by making an HTTP GET request and parsing the response body.
func GrabDownloadUrl(url string) (string, error) {
	resp, err := httpclient.Client.Get(url)
	if err != nil {
		return consts.EMPTY_STRING, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return consts.EMPTY_STRING, err
	}

	var data types.GitHubResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return consts.EMPTY_STRING, err
	}

	return data.DownloadURL, nil
}

// SaveFile: Downloads a file from the specified URL and saves it to the provided output path. It creates any necessary directories, handles overwriting existing files, and uses a spinner for user feedback.
// SaveFile downloads a file from the specified URL and saves it to the outputPath.
// It shows a spinner during the operation and handles interruptions.
func SaveFile(url, outputPath string, overwrite bool) error {
	// Create a context with cancellation
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a spinner with the DefaultSpinnerFactory
	s, err := spinner.CreateSpinner()
	if err != nil {
		return err
	}

	// Ensure the spinner stops when the function exits
	defer s.Stop()

	// Setup StopOnSignal to handle interruptions
	spinner.StopOnSignal(s)

	// Start the spinner
	if err := s.Start(); err != nil {
		return err
	}

	// Show initial processing message
	file := filepath.Base(outputPath)
	dir := filepath.Dir(outputPath)
	s.Message(fmt.Sprintf("Processing %s", color.New(color.FgCyan).Sprint(file)))

	// Check if file exists and whether to overwrite it
	if _, err := os.Stat(outputPath); err == nil && !overwrite {
		s.StopMessage(fmt.Sprintf("Skipped %s", color.New(color.FgRed).Sprint(file)))
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// Create the directory structure if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to create directory structure for '%s': %v", outputPath, err))
		return fmt.Errorf("failed to create directory structure for '%s': %v", outputPath, err)
	}

	// Download the file
	resp, err := httpclient.Client.Get(url)
	if err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to get contents from %s: %v", url, err))
		return fmt.Errorf("failed to get contents from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		s.StopFailMessage(fmt.Sprintf("failed to get contents %s: HTTP status %d", url, resp.StatusCode))
		return fmt.Errorf("failed to get contents %s: HTTP status %d", url, resp.StatusCode)
	}

	// Create the output file
	out, err := os.Create(outputPath)
	if err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to create file '%s': %v", outputPath, err))
		return fmt.Errorf("failed to create file '%s': %v", outputPath, err)
	}
	defer out.Close()

	// Copy response body to the file
	if _, err := io.Copy(out, resp.Body); err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to write file '%s': %v", outputPath, err))
		return fmt.Errorf("failed to write file '%s': %v", outputPath, err)
	}

	// Show success message
	time.Sleep(500 * time.Millisecond)
	s.StopMessage(fmt.Sprintf("Processed %s saved in %s", color.New(color.FgGreen).Sprint(file), color.New(color.FgCyan).Sprint(dir)))

	return nil
}

// GetReleaseFile: Returns the appropriate release file for the given programming language. Currently, it supports Go and returns an error for unsupported languages.
func GetReleaseFile(projectLanguage string) (string, error) {
	if projectLanguage == consts.EMPTY_STRING {
		return consts.EMPTY_STRING, nil
	}

	switch strings.ToLower(projectLanguage) {
	case consts.GO_LANG:
		return consts.GORELEASER, nil
	default:
		return consts.EMPTY_STRING, fmt.Errorf("release file for projectLanguage: %s hasn't been implemented yet", projectLanguage)
	}
}

// GetVersionFile: Returns the version file path for the specified project language. If the language is unsupported, it returns an error.
func GetVersionFile(projectLanguage string) (string, error) {
	if projectLanguage == consts.EMPTY_STRING {
		return consts.EMPTY_STRING, nil
	}

	switch strings.ToLower(projectLanguage) {
	case consts.GO_LANG:
		return consts.VERSION_GO, nil
	default:
		return consts.EMPTY_STRING, fmt.Errorf("version file for projectLanguage: %s hasn't been implemented yet", projectLanguage)
	}
}
