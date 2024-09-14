package types

// GitHubResponse represents a simplified response from the GitHub API containing the download URL.
type GitHubResponse struct {
	DownloadURL string `json:"download_url"`
}

// GitHubItem represents an item in a GitHub repository, including its type, name, path, and download URL.
type GitHubItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

// CliFlags holds the flags passed by the user in the CLI, such as repository information and configuration options.
type CliFlags struct {
	BranchName         string
	IncludeMakefile    bool
	IncludeVersionFile bool
	LicenseType        string
	OutputDirectory    string
	OverwriteFiles     bool
	ProjectLanguage    string
	GithubToken        string
	RepoOwner          string
	RepoName           string
}
