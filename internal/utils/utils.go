package utils

import (
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
func SaveFile(url, outputPath string, overwrite bool) error {
	s, err := spinner.CreateSpinner()
	if err != nil {
		return err
	}

	spinner.StopOnSignal(s)

	if err := s.Start(); err != nil {
		return err
	}

	defer s.Stop()

	file := filepath.Base(outputPath)
	dir := filepath.Dir(outputPath)

	s.Message(fmt.Sprintf("Processing %s", SetColor(color.LightCyan, file)))

	if _, err := os.Stat(outputPath); err == nil && !overwrite {
		s.StopMessage(fmt.Sprintf("Skipped %s", SetColor(color.FgLightRed, file)))
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to create directory structure for '%s': %v", outputPath, err))
		return fmt.Errorf("failed to create directory structure for '%s': %v", outputPath, err)
	}

	resp, err := httpclient.Client.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.StopFailMessage(fmt.Sprintf("failed to get contents %s: HTTP status %d", url, resp.StatusCode))
		return fmt.Errorf("failed to get contents %s: HTTP status %d", url, resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to create file '%s': %v", outputPath, err))
		return fmt.Errorf("failed to create file '%s': %v", outputPath, err)
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		s.StopFailMessage(fmt.Sprintf("failed to write file '%s': %v", outputPath, err))
		return fmt.Errorf("failed to write file '%s': %v", outputPath, err)
	}

	time.Sleep(500 * time.Millisecond)
	s.StopMessage(fmt.Sprintf("Processed %s saved in %s", SetColor(color.FgLightGreen, file), SetColor(color.FgCyan, dir)))

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
