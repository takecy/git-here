# git-here

![unittest](https://github.com/takecy/git-here/workflows/unittest/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/takecy/git-here)](https://goreportcard.com/report/github.com/takecy/git-here)
![Go Version](https://img.shields.io/badge/golang-1.24+-blue.svg?style=flat-square)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/takecy/git-here)

**Efficiently run git commands across multiple repositories in parallel**

**Before:**
```bash
cd project1 && git pull && cd ..
cd project2 && git pull && cd ..
cd project3 && git pull && cd ..
# ... repeat for dozens of repositories
```

**After:**
```bash
gih pull  # Done! All repositories updated in parallel
```

It is just a tool to do this. It does nothing else.  
I created it because I was tired of managing dozens of microservice repositories for the projects I work on.


## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Examples](#examples)
- [Command Reference](#command-reference)
- [Development](#development)
- [Contributing](#contributing)
- [FAQ](#faq)
- [License](#license)


## Features

- 🚀 **Parallel Execution**: Execute git commands across multiple repositories simultaneously
- 🎯 **Filtering**: Use regex patterns to target specific repositories or exclude others
- 🔍 **Directory Discovery**: Automatically finds all git repositories in subdirectories
- ⚡  **High Performance**: Leverages Go's goroutines for maximum efficiency
- ⏰ **Configurable Timeouts**: Aborts immediately at the specified timeouts


<br/>

## Installation

### via Go Install (Recommended)

```bash
go install github.com/takecy/git-here/gih@latest
```

**Verify installation:**
```bash
gih version
```

### via Binary Download

1. Download the latest binary from the [Release Page](https://github.com/takecy/git-here/releases) for your platform
2. Extract and place the binary in your `$PATH`
3. Make it executable: `chmod +x gih` (Unix-like systems)

**Supported Platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64) 
- Windows (amd64, arm64)

## Usage

### Basic Syntax

```bash
gih [options] <git_command> [git_options]
```

### Quick Start

```bash
# Pull all repositories in current directory
gih pull

# Check status of all repositories  
gih status

# Fetch from all remotes
gih fetch --all
```

## Examples

### Common Operations

**Update all repositories:**
```bash
gih pull
# Equivalent to running 'git pull' in each repository directory
```

**Check status across all repositories:**
```bash
gih status --short
# Shows git status for each repository in a compact format
```

**Fetch from all remotes:**
```bash
gih fetch --all --prune
# Fetches from all remotes and prunes deleted branches
```

### Advanced Filtering

**Target specific repositories with regex:**
```bash
# Only operate on repositories matching the pattern
gih --target "^(frontend|backend)" pull
# Only pulls repositories starting with 'frontend' or 'backend'

gih --target ".*-service$" status  
# Only checks status of repositories ending with '-service'
```

**Exclude repositories:**
```bash
# Ignore specific repositories
gih --ignore "^(test|temp)" pull
# Pulls all repositories except those starting with 'test' or 'temp'

gih --ignore ".*-wip$" status
# Check status of all repositories except those ending with '-wip'
```

**Combine target and ignore patterns:**
```bash
gih --target "^microservice" --ignore "deprecated" pull
# Pull only microservice repositories that aren't deprecated
```

### Performance Tuning

**Set custom timeout:**
```bash
# Increase timeout for slow operations
gih --timeout 60s pull
# Each repository operation will timeout after 60 seconds

gih --timeout 5m clone --recursive
# For operations that might take longer
```

**Control concurrency:**
```bash
# Limit concurrent operations (useful for resource-constrained environments)
gih -c 2 pull
# Only run 2 operations in parallel

# Maximize parallelism (default is number of CPU cores)  
gih -c 20 fetch
# Run up to 20 operations in parallel
```

### Real-World Scenarios

**Microservices maintenance:**
```bash
# Update all microservices before deployment
gih --target ".*-service$" pull

# Check which microservices have uncommitted changes
gih --target ".*-service$" status --porcelain
```

**Open source contribution workflow:**
```bash
# Fetch latest changes from all your forks
gih fetch upstream

# Check status of all your projects
gih status --short
```

**Monorepo-adjacent development:**
```bash
# Update all related projects in your workspace
gih --ignore "node_modules|\.venv" pull
```

### Directory Structure

git-here discovers repositories based on this structure:

```
workspace/
├── .hidden-repo/     # ❌ Ignored (hidden directories starting with .)
├── project-a/        # ✅ Target (contains .git)
│   └── .git/
├── project-b/        # ✅ Target (contains .git)  
│   └── .git/
├── non-git-dir/      # ❌ Ignored (no .git directory)
│   └── some-file.txt
└── nested/
    └── deep-repo/    # ❌ Not target (only scans direct subdirectories)
        └── .git/
```

**Key Rules:**
- Only direct subdirectories are scanned (not deeply nested)
- Directories must contain a `.git` folder to be considered repositories
- Hidden directories (starting with `.`) are automatically ignored

## Command Reference

### Options

| Option | Description | Default | Example |
|--------|-------------|---------|----------|
| `--target` | Regex pattern to target specific directories | `""` (all) | `--target "^api"` |
| `--ignore` | Regex pattern to ignore directories | `""` (none) | `--ignore "test$"` |
| `--timeout` | Timeout per repository operation | `20s` | `--timeout 60s` |
| `-c` | Concurrency level (max parallel operations) | CPU cores | `-c 4` |

### Commands

**Built-in:**
- `gih version` - Show version information and check for updates
- `gih` - Show help message

**Git Commands:**
Any valid git command can be used:
- `gih pull` - Pull latest changes
- `gih fetch --all` - Fetch from all remotes
- `gih status --short` - Show compact status
- `gih push origin main` - Push to origin/main
- `gih checkout -b feature/new` - Create and checkout new branch
- `gih commit -m "message"` - Commit changes
- And many more...

### Exit Codes

- `0` - Success (all operations completed)
- `1` - General error (invalid arguments, setup issues)
- `2` - Some repositories failed (partial success)

## Development

### Prerequisites

- Go 1.24 or later

### Setup

```bash
# Clone the repository
git clone https://github.com/takecy/git-here.git
cd git-here

# Build for development
make build

# Run tests
make test

# Run linter
make lint
```

### Available Make Commands

```bash
make build      # Build development binary (gih_dev)
make install    # Install production binary
make test       # Run all tests with race detection
make lint       # Run golangci-lint
make tidy       # Run go mod tidy
make update     # Update all dependencies
```

### Development Workflow

```bash
# Build and test your changes
make build
make test
make lint

# Test manually with debug output
DEBUG=* go run gih/main.go status

# Test with the development binary
./gih_dev status
```

### Architecture

The project is structured into several key packages:

- **`gih/main.go`** - CLI entry point and argument parsing
- **`syncer/`** - Core orchestration logic
  - `syncer.go` - Main execution controller  
  - `dir.go` - Repository discovery
  - `git.go` - Git command execution
- **`printer/`** - Output formatting and coloring

## Contributing

We welcome contributions! Here's how to get started:

### Reporting Issues

- Use the [GitHub Issues](https://github.com/takecy/git-here/issues) page
- Include your OS, Go version, and git-here version
- Provide steps to reproduce the issue
- Include relevant command output

### Pull Requests

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes with tests
4. **Run** the full test suite: `make test lint`
5. **Commit** your changes: `git commit -m 'Add amazing feature'`
6. **Push** to your branch: `git push origin feature/amazing-feature`
7. **Open** a Pull Request

### Code Style

- Follow standard Go conventions
- Run `make lint` before submitting
- Add tests for new functionality
- Update documentation as needed

### Testing

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./syncer/

# Run with coverage
go test -cover ./...
```

## FAQ

### Q: Why doesn't git-here work with deeply nested repositories?
**A:** By design, git-here only scans direct subdirectories for performance and clarity. If you need to work with deeply nested repos, consider running git-here from different directory levels.

### Q: Can I use git-here with non-git commands?
**A:** No, git-here is specifically designed for git commands. It expects git repositories and uses git-specific logic.

### Q: How do I handle repositories that require different authentication?
**A:** git-here uses your existing git configuration and SSH keys. Ensure your git setup works normally first.

### Q: What happens if one repository fails?
**A:** Individual repository failures don't stop the entire operation. Failed repositories are reported, but git-here continues with the remaining repositories.

### Q: Can I use environment variables for configuration?
**A:** Currently, git-here uses command-line flags. Environment variable support may be added in future versions.

### Q: How do I update git-here?
**A:** Run `go install github.com/takecy/git-here/gih@latest` or download the latest binary from the releases page.

## License

[MIT](./LICENSE) © [takecy](https://github.com/takecy)