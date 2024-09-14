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

// FileOpsInterface defines methods for file operations
type FileOpsInterface interface {
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
}

// FileOps is a real implementation that uses os package functions
type FileOps struct{}

func (f *FileOps) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *FileOps) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func SetColor(col color.Color, item interface{}) string {
	return col.Sprintf("%v", item)
}

func GrabDownloadUrl(url string) (string, error) {
	if httpclient.Client == nil {
		return consts.EMPTY_STRING, fmt.Errorf("HTTP client is not initialized")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return consts.EMPTY_STRING, err
	}

	resp, err := httpclient.Client.Do(req)
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

func SaveFile(url, outputPath string, overwrite bool) error {
	return SaveFileWithSpinner(url, outputPath, overwrite, spinner.CreateSpinner, &FileOps{})
}

func SaveFileWithSpinner(url, outputPath string, overwrite bool, spinnerCreator func() (spinner.SpinnerInterface, error), fileOps FileOpsInterface) error {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a spinner with the DefaultSpinnerFactory
	s, err := spinnerCreator()
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
	if _, err := fileOps.Stat(outputPath); err == nil && !overwrite {
		s.StopMessage(fmt.Sprintf("Skipped %s", color.New(color.FgRed).Sprint(file)))
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// Create the directory structure if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to create directory structure for '%s': %v", outputPath, err))
		return fmt.Errorf("failed to create directory structure for '%s': %v", outputPath, err)
	}

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to create request for %s: %v", url, err))
		return fmt.Errorf("failed to create request for %s: %v", url, err)
	}

	// Download the file
	resp, err := httpclient.Client.Do(req)
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
	out, err := fileOps.Create(outputPath)
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
