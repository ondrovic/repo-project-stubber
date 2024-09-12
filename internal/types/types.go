package types

type GitHubItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

type CliFlags struct {
	BranchName      string
	IncludeMakefile bool
	LicenseType     string
	OutputDirectory string
	OverwriteFiles  bool
	ProjectLanguage string
	GithubToken     string
	RepoOwner       string
	RepoName        string
}