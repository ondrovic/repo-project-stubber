![License](https://img.shields.io/badge/license-MIT-blue)
[![testing](https://github.com/ondrovic/repo-project-stubber/actions/workflows/testing.yml/badge.svg)](https://github.com/ondrovic/repo-project-stubber/actions/workflows/testing.yml)
[![goreleaser](https://github.com/ondrovic/repo-project-stubber/actions/workflows/releaser.yml/badge.svg)](https://github.com/ondrovic/repo-project-stubber/actions/workflows/releaser.yml)

# Repo Project Stubber

A CLI tool to create project stubs with templated files based on user-defined options.

## Features

- Download GitHub repository contents when creating a new project
- Customizable project setup with various options
- Support for multiple programming languages
- Automatic creation of common project files (README, LICENSE, .gitignore, etc.)
- Configurable workflow files for GitHub Actions
- Optional inclusion of Makefile and version file
- Clear terminal screen functionality

## Important Information

This CLI tool works in conjunction with [Vscode template repo](https://github.com/ondrovic/vscode) which is where the files are pulled from.

## Usage

```bash
repo-stub <output-directory> [flags]
```
### Flags

- `-r, --repo-name string`: Name of the repository (default "vscode")
- `-o, --repo-owner string`: Owner of the repository (default "ondrovic")
- `-b, --branch-name string`: Branch name you wish to pull from (default "master")
- `-t, --github-token string`: GitHub API token
- `-p, --project-language string`: What language is your app in (default "go")
- `-l, --license-type string`: What license are you using (default "mit")
- `-m, --include-makefile`: Include a Makefile
- `-v, --include-version-file`: Include a version file
- `-w, --overwrite-files`: Overwrite existing files

## Examples

Create a new Go project with MIT license:

```bash
repo-stub my-new-project -p go -l mit
```

Create a Python project with a Makefile and version file:

```bash
repo-stub my-python-project -p python -m -v
```
## Version Command

The CLI includes a version command that provides information about the current version and checks for available upgrades:

```bash
repo-stub version
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.