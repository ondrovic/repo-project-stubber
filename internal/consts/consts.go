package consts

// Programming language abbreviations and their full forms used in the project.
const (
	// GO_LANG represents the abbreviation for Go programming language.
	GO_LANG = "go"

	// PY_LANG represents the abbreviation for Python programming language.
	PY_LANG = "py"

	// PYTHON represents the full form of the Python programming language.
	PYTHON = "python"

	// JS_LANG represents the abbreviation for JavaScript programming language.
	JS_LANG = "js"

	// TS_LANG represents the abbreviation for TypeScript programming language.
	TS_LANG = "ts"

	// JAVASCRIPT represents the full form of the JavaScript programming language.
	JAVASCRIPT = "javascript"

	// TYPESCRIPT represents the full form of the TypeScript programming language.
	TYPESCRIPT = "typescript"
)

// File and directory types and special filenames used across the project.
const (
	// DIR_TYPE represents a directory type identifier.
	DIR_TYPE = "dir"

	// FILE_TYPE represents a file type identifier.
	FILE_TYPE = "file"

	// GIT_HUB represents the directory name for the github files.
	GIT_HUB = ".github"

	// GIT_IGNORE represents the filename for .gitignore.
	GIT_IGNORE = ".gitignore"

	// GIT_KEEP represents the filename for .gitkeep.
	GIT_KEEP = ".gitkeep"

	// GORELEASER represents the filename for the go release file.
	GORELEASER = "goreleaser.yaml"

	// LICENSE represents the filename for license files.
	LICENSE = "LICENSE"

	// MAKEFILE represents the filename for the makefile files.
	MAKEFILE = "Makefile"

	// README represents the filename for README files.
	README = "README.md"

	//-TODO represents the filename for TODO files.
	TODO = "TODO"

	// VSCODE represents the directory name for the VSCODE files.
	VSCODE = ".vscode"

	// VERSION_GO represents the filename for the version file for go.
	VERSION_GO = "version.go"

	// WORKFLOW represents the directory name for the workflow files.
	WORKFLOW = "workflows"

	// YML represent the extension for teh yml files.
	YML = ".yml"
)

// Categories of files that are typically ignored, licensed, or related to project configuration.
const (
	// IGNORE_FILES is used to represent files that should be ignored.
	IGNORE_FILES = ".ignoreFiles"

	// LICENSE_FILES represents the files related to licensing information.
	LICENSE_FILES = ".licenseFiles"

	// MAKE_FILES represents the files related to build scripts, e.g., Makefile.
	MAKE_FILES = ".makeFiles"

	// README_FILES represents files related to README documentation.
	README_FILES = ".readmeFiles"

	// RELEASE_FILES represents files related to release processes.
	RELEASE_FILES = ".releaseFiles"

	// TODO_FILES represents files related to TODO tracking.
	TODO_FILES = ".todoFiles"

	// VERSION_FILES represents files related to versioning.
	VERSION_FILES = ".versionFiles"

	// VSCODE_FILES represents files related to Visual Studio Code configuration.
	VSCODE_FILES = ".vscodeFiles"

	// WORKFLOW_FILES represents files related to CI/CD workflows.
	WORKFLOW_FLIES = ".workflowFiles"
)
