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
	// "github.com/gookit/color"
)

func FileDownloader(url, outputPath string, overwrite bool) error {

	item, err := GetFileInfo(url)
	if err != nil {
		return fmt.Errorf("failed to get file info %s: %v", url, err)
		// return err
	}

	return SaveFile(item.DownloadURL, outputPath, overwrite)
}

func GetFileInfo(url string) (*types.GitHubItem, error) {
	resp, err := httpclient.Client.Get(url)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("failed to get file info from '%s': %v", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get contents %s: HTTP status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var items []types.GitHubItem
	err = json.Unmarshal(body, &items)
	if err == nil && len(items) > 0 {
		// If it's an array, return the first item
		return &items[0], nil
	}

	// If it's not an array, try to unmarshal as a single item
	var item types.GitHubItem
	err = json.Unmarshal(body, &item)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("failed to decode file info: %v", err)
	}

	return &item, nil
}

// func SetColor(col color.Color, item string) string {
// 	return col.Sprintf(item)
// }

func SetColor(col color.Color, item interface{}) string {
	return col.Sprintf("%v", item)
}
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
	// dir := filepath.Dir(outputPath)

	s.Message(fmt.Sprintf("Processing %s", SetColor(color.LightCyan, file)))

	if _, err := os.Stat(outputPath); err == nil && !overwrite {
		// s.StopMessage(fmt.Sprintf("Skipped %s in %s", SetColor(color.FgGreen, file), SetColor(color.FgLightCyan, dir)))
		s.StopMessage(fmt.Sprintf("Skipped %s", SetColor(color.FgLightRed, file)))
		time.Sleep(1 * time.Second)
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
		// return nil
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

	time.Sleep(1 * time.Second)
	// s.StopMessage(fmt.Sprintf("Created %s in %s", SetColor(color.FgLightGreen, file), SetColor(color.FgLightCyan, dir)))
	s.StopMessage(fmt.Sprintf("Processed %s", SetColor(color.FgLightGreen, file)))

	return nil
}

func GetReleaseFile(projectLanguage string) (string, error) {
	if projectLanguage == "" {
		return "", nil
	}

	switch strings.ToLower(projectLanguage) {
	case consts.GO_LANG:
		return consts.GORELEASER, nil
	default:
		return "", fmt.Errorf("release file for projectLanguage: %s hasn't been implemented yet", projectLanguage)
	}
}

func GetVersionFile(projectLanguage string) (string, error) {
	if projectLanguage == "" {
		return "", nil
	}

	switch strings.ToLower(projectLanguage) {
	case consts.GO_LANG:
		return consts.VERSION_GO, nil
	default:
		return "", fmt.Errorf("version file for projectLanguage: %s hasn't been implemented yet", projectLanguage)
	}
}
