package types

type GitHubResponse struct {
	DownloadURL string `json:"download_url"`
}

type GitHubItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

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
