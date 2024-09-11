package utils

import (
	"encoding/json"
	"fmt"
	"github-project-template/internal/consts"
	"github-project-template/internal/httpclient"
	"github-project-template/internal/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func FileDownloader(url, outputPath string, overwrite bool) error {

	item, err := GetFileInfo(url)
	if err != nil {
		// return fmt.Errorf("failed to get file info %s: %v", url, err)
		return err
	}
	// TODO: do a choice if you want to overwrite, this will give you the ability to only overwrite certain files

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

func SaveFile(url, outputPath string, overwrite bool) error {
	// todo: spinner here
	if _, err := os.Stat(outputPath); err == nil && !overwrite {
		// todo notify  user of skipping
		fmt.Printf("Skipping '%s'\n", outputPath)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		// todo: notify user
		// return err
		return fmt.Errorf("failed to create directory structure for '%s': %v", outputPath, err)
	}

	// spinner
	// resp, err := http.Get(url)
	resp, err := httpclient.Client.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get contents %s: HTTP status %d", url, resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %v", outputPath, err)
		// return err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file '%s': %v", outputPath, err)
		// return err
	}

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
