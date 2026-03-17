# SyncX - Repository Synchronization Assistant

A modern, intelligent repository synchronization tool for managing multiple Git projects with smart tracking, automatic directory management, and GitLab API integration.

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd syncx

# Install globally using make
make install

# Or use the install script directly
./scripts/install.sh
```

After installation, the `syncx` command will be available globally:
```bash
syncx --help
syncx clone --file projects-inventory.json --protocol ssh -o ~/repos
syncx pull --file projects-inventory.json -o ~/repos
```

### Manual Installation

```bash
# Build the binary
go build -o syncx main.go

# Copy to your PATH (choose one)
sudo cp syncx /usr/local/bin/
# OR
cp syncx ~/bin/  # Make sure ~/bin is in your PATH
```

### Uninstall

```bash
make uninstall
# OR
./scripts/uninstall.sh
```

---

## 📋 Commands Overview

| Command | Purpose | Best For |
|---------|---------|----------|
| `generate-json` | Generate inventory from GitLab API | Initial setup, inventory refresh |
| `clone` | Clone new + update existing | Daily sync, full repository management |
| `pull` | Update existing projects only | Quick updates without new clones |
| `check` | Check for uncommitted local changes | Pre-sync validation, change detection |
| `scan` | Recursively scan directory for git repos | No inventory needed, workspace scanning |
| `list` | Show projects and groups | Discovery, validation |
| `status` | Check repository status | Monitoring, troubleshooting |

---

## 🔧 `generate-json` — Generate Inventory from GitLab

Connects to the GitLab API and generates a `projects-inventory.json` file that reflects the full group/project hierarchy of your GitLab organization. The generated file includes metadata for each project (`default_branch`, `description`, `http_url`) to enable smarter operations.

```bash
syncx generate-json --token <gitlab-token> --group <group-path> --out <output-dir>
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--token` | GitLab Personal Access Token | required |
| `--group` | GitLab group path (e.g. `myorg/mygroup`) | required |
| `--out` | Output directory for the generated file | `.` |
| `--host` | GitLab host (for self-hosted instances) | `gitlab.com` |
| `--filename` | Output filename | `projects-inventory.json` |

### Examples

```bash
# Generate inventory for a GitLab group
syncx generate-json --token glpat-xxxx --group myorg/mygroup --out ~/

# Specify a custom output filename
syncx generate-json --token glpat-xxxx --group myorg/mygroup --out ~/ --filename my-inventory.json

# Self-hosted GitLab instance
syncx generate-json --token glpat-xxxx --group myorg/mygroup --out ~/ --host gitlab.mycompany.com
```

### Output

The command generates a hierarchical JSON file compatible with all `syncx` commands:

```
✅ Inventory generated successfully!
   📁 File:     ~/projects-inventory.json
   📦 Groups:   12
   🗂️  Projects: 87
```

---

## 🚀 `clone` — Clone and Update Repositories

Clones new repositories and updates existing ones based on your inventory file. Intelligently uses `default_branch` from the inventory to ensure the correct branch is checked out.

```bash
syncx clone --file projects-inventory.json --protocol ssh -o ~/repos
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--file` | Path to inventory JSON file | `projects-inventory.json` |
| `--protocol` | Git protocol: `ssh` or `http` | `ssh` |
| `-o, --output` | Output directory for repositories | auto-detected |
| `--group` | Filter operations to a specific group | all groups |
| `--parallel` | Number of parallel operations | 1 |
| `--dry-run` | Preview operations without executing | false |
| `-v, --verbose` | Verbose output | false |
| `--show-groups` | Show available groups and exit | false |

### Examples

```bash
# Clone all repositories (new) and update existing ones
syncx clone --file projects-inventory.json --protocol ssh -o ~/repos

# Preview what would happen without executing
syncx clone --file projects-inventory.json --protocol ssh -o ~/repos --dry-run -v

# Clone only a specific group
syncx clone --file projects-inventory.json --protocol ssh -o ~/repos --group "Backend"

# Speed up with parallel processing
syncx clone --file projects-inventory.json --protocol ssh -o ~/repos --parallel 8

# Use HTTPS instead of SSH
syncx clone --file projects-inventory.json --protocol http -o ~/repos

# Show all available groups
syncx clone --file projects-inventory.json --show-groups
```

---

## 🔄 `pull` — Update Existing Repositories

Updates only already-cloned repositories. Does not clone new projects.

```bash
syncx pull --file projects-inventory.json -o ~/repos
```

### Flags

Same as `clone`, except it never clones new repositories.

### Examples

```bash
# Update all existing repositories
syncx pull --file projects-inventory.json -o ~/repos

# Update with verbose output
syncx pull --file projects-inventory.json -o ~/repos -v

# Update only Backend group
syncx pull --file projects-inventory.json -o ~/repos --group "Backend"

# Update in parallel
syncx pull --file projects-inventory.json -o ~/repos --parallel 5
```

### `clone` vs `pull` comparison

| Feature | `clone` | `pull` |
|---------|---------|--------|
| Clone new projects | ✅ | ❌ |
| Update existing | ✅ | ✅ |
| Smart tracking | ✅ | ✅ |
| Group filtering | ✅ | ✅ |
| Parallel processing | ✅ | ✅ |

---

## 🔍 `check` — Check for Uncommitted Changes

Scans repositories for uncommitted local changes before syncing, so you don't accidentally overwrite work in progress.

```bash
syncx check --file projects-inventory.json -o ~/repos
```

### Flags

| Flag | Description |
|------|-------------|
| `--file` | Path to inventory JSON file |
| `-o, --output` | Directory where repositories are cloned |
| `--group` | Filter to a specific group |
| `--parallel` | Number of parallel checks |
| `-v, --verbose` | Show clean repositories too |

### Examples

```bash
# Check all repositories for uncommitted changes
syncx check --file projects-inventory.json -o ~/repos

# Verbose: show both dirty and clean repositories
syncx check --file projects-inventory.json -o ~/repos -v

# Check only a specific group
syncx check --file projects-inventory.json -o ~/repos --group "Frontend"

# Fast parallel check
syncx check --file projects-inventory.json -o ~/repos --parallel 20
```

---

## 📡 `scan` — Scan Directories Without Inventory

Recursively scans a directory for git repositories with uncommitted changes. No inventory file needed.

```bash
syncx scan ~/repos
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `path` | Directory to scan (positional argument) | `.` |
| `-d, --depth` | Maximum directory recursion depth | 5 |
| `--parallel` | Number of parallel operations | 10 |
| `--show-clean` | Also show repositories without changes | false |
| `-v, --verbose` | Show full repository paths | false |

### Examples

```bash
# Scan current directory
syncx scan .

# Scan a specific directory
syncx scan ~/repos

# Limit scan depth for faster results
syncx scan ~/workspace -d 3

# Show all repositories, including clean ones
syncx scan ~/repos --show-clean

# Scan with full paths in output
syncx scan ~/projects -v

# Fast parallel scan of a large workspace
syncx scan ~/workspace --parallel 20 -d 4

# Scan multiple locations
syncx scan ~/production-repos
syncx scan ~/dev-repos

# Scan home directory for forgotten repositories
syncx scan ~ -d 4
```

---

## 📋 `list` — List Projects from Inventory

Displays all groups and projects from your inventory file. Shows `default_branch` next to each project name and `description` when available.

```bash
syncx list --file projects-inventory.json
```

### Flags

| Flag | Description |
|------|-------------|
| `--file` | Path to inventory JSON file |
| `--groups-only` | Show only groups, not individual projects |
| `--compact` | Compact display (one project per line) |
| `-v, --verbose` | Show additional details (local path, etc.) |

### Examples

```bash
# Full detailed view (shows branch and description)
syncx list --file projects-inventory.json

# Compact view (one project per line)
syncx list --file projects-inventory.json --compact

# Show only groups summary
syncx list --file projects-inventory.json --groups-only

# Verbose: includes local paths
syncx list --file projects-inventory.json -v
```

### Sample Output

```
📁 Backend (3 projects)
  📦 api-server  [main]
      💬 REST API for the main application
      🔗 git@gitlab.com:myorg/backend/api-server.git

  📦 auth-service  [develop]
      🔗 git@gitlab.com:myorg/backend/auth-service.git
```

---

## 📊 `status` — Check Repository Status

Shows the health and sync status of cloned repositories.

```bash
syncx status --file projects-inventory.json -o ~/repos
```

### Examples

```bash
# Check status of all repositories
syncx status --file projects-inventory.json -o ~/repos

# Verbose status with details
syncx status --file projects-inventory.json -o ~/repos -v
```

---

## ⚙️ Global Flags

These flags are available on all commands:

| Flag | Description | Default |
|------|-------------|---------|
| `--file` | Path to inventory JSON file | `projects-inventory.json` |
| `--protocol` | Git protocol: `ssh` or `http` | `ssh` |
| `-o, --output` | Base output directory | auto-detected |
| `-v, --verbose` | Enable verbose output | false |
| `--dry-run` | Preview operations without executing | false |

---

## 🎯 Common Workflows

### First-time setup from GitLab

```bash
# 1. Generate the inventory from your GitLab group
syncx generate-json --token glpat-xxxx --group myorg/mygroup --out ~/

# 2. Clone all repositories
syncx clone --file ~/projects-inventory.json --protocol ssh -o ~/repos -v
```

### Daily sync workflow

```bash
# 1. Check for uncommitted changes first
syncx check --file projects-inventory.json -o ~/repos -v

# 2. Pull updates (or clone new + update existing)
syncx pull --file projects-inventory.json -o ~/repos -v
```

### Refresh inventory and sync new projects

```bash
# 1. Regenerate inventory (picks up new repos from GitLab)
syncx generate-json --token glpat-xxxx --group myorg/mygroup --out ~/

# 2. Clone new + update existing
syncx clone --file ~/projects-inventory.json --protocol ssh -o ~/repos -v
```

### Work with a specific team's group

```bash
# Clone only the Frontend group
syncx clone --file projects-inventory.json --protocol ssh -o ~/repos --group "Frontend" -v

# Update only Backend
syncx pull --file projects-inventory.json -o ~/repos --group "Backend" -v

# Check only DevOps for uncommitted changes
syncx check --file projects-inventory.json -o ~/repos --group "DevOps"
```

### Managing multiple environments

```bash
# Production clone
syncx clone --file projects-inventory.json --protocol ssh -o ~/production-repos
syncx check --file projects-inventory.json -o ~/production-repos

# Development clone
syncx clone --file projects-inventory.json --protocol ssh -o ~/dev-repos
syncx check --file projects-inventory.json -o ~/dev-repos

# Backup
syncx pull --file projects-inventory.json -o ~/backup-repos
```

### End-of-day check (no inventory needed)

```bash
# Find any repository with uncommitted changes across your whole workspace
syncx scan ~/workspace -v
syncx scan ~/projects -d 3
```

---

## 📋 Inventory File Format

The `projects-inventory.json` file uses a hierarchical format with nested groups. Fields marked as optional are populated automatically when using `generate-json`.

```json
{
  "groups": [
    {
      "name": "Backend",
      "description": "Server-side services",
      "projects": [
        {
          "name": "api-server",
          "url": "git@gitlab.com:myorg/backend/api-server.git",
          "http_url": "https://gitlab.com/myorg/backend/api-server.git",
          "default_branch": "main",
          "description": "Main REST API"
        },
        {
          "name": "auth-service",
          "url": "git@gitlab.com:myorg/backend/auth-service.git",
          "http_url": "https://gitlab.com/myorg/backend/auth-service.git",
          "default_branch": "develop"
        }
      ],
      "groups": [
        {
          "name": "Microservices",
          "projects": [
            {
              "name": "user-service",
              "url": "git@gitlab.com:myorg/backend/microservices/user-service.git",
              "default_branch": "main"
            }
          ]
        }
      ]
    },
    {
      "name": "DevOps",
      "skip": true,
      "projects": [
        {
          "name": "infrastructure",
          "url": "git@gitlab.com:myorg/devops/infrastructure.git",
          "default_branch": "main"
        }
      ]
    }
  ],
  "projects": [
    {
      "name": "documentation",
      "url": "git@gitlab.com:myorg/documentation.git",
      "default_branch": "main"
    }
  ]
}
```

### Project Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | ✅ | Project name |
| `url` | ✅ | SSH clone URL |
| `http_url` | optional | HTTPS clone URL (used with `--protocol http`) |
| `default_branch` | optional | Branch to checkout on clone |
| `description` | optional | Shown in `syncx list` output |

### Group Fields

| Field | Description |
|-------|-------------|
| `name` | Group name |
| `description` | Group description |
| `skip` | Set to `true` to exclude this group from all operations |
| `projects` | List of projects in this group |
| `groups` | Nested subgroups |

---

## 📁 Project Structure

```
syncx/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── README.md               # Complete usage documentation
├── CLAUDE.md               # Development guidance for Claude Code
│
├── cmd/                    # Cobra CLI commands
│   ├── root.go             # Root command and global flags
│   ├── clone.go            # Clone/update repositories with smart tracking
│   ├── pull.go             # Update existing repositories only
│   ├── check.go            # Check for uncommitted changes
│   ├── scan.go             # Scan directories for git repos (no inventory)
│   ├── list.go             # List projects and groups with metadata
│   ├── status.go           # Check repository status
│   └── generate_json.go    # Generate inventory from GitLab API
│
├── internal/               # Core functionality
│   ├── types.go            # Data structures and types
│   ├── logger.go           # Colored logging system
│   ├── git.go              # Git operations (clone with branch support)
│   ├── inventory.go        # Inventory file processing
│   └── tracker.go          # Smart repository tracking system
│
└── scripts/                # Build and utility scripts
    ├── build.sh             # Build script
    ├── install.sh           # Installation script
    └── uninstall.sh         # Uninstall script
```

---

## ✨ Key Features

### 🔧 GitLab API Integration
- **Automatic inventory generation** from any GitLab group or subgroup
- **Recursive group traversal** capturing the full project hierarchy
- **Rich metadata** (`default_branch`, `description`, `http_url`) stored in inventory
- **Self-hosted GitLab** support via `--host` flag
- **Paginated API** handling for large organizations

### 🎯 Smart Tracking System
- **Automatic directory creation** for missing group structures
- **Intelligent diff detection** between inventory JSON and physical structure
- **MD5-based change detection** for inventory updates
- **Persistent tracking** with `.syncx-tracker.json` files
- **Git change detection** using `git fetch` to check for remote updates

### 🚀 Enhanced Git Operations
- **Branch-aware cloning** — uses `default_branch` from inventory to checkout the correct branch
- **Robust clone operations** with verification and error handling
- **Smart pull strategy** with fetch-first approach
- **Conflict resolution** with fallback mechanisms
- **SSH and HTTPS protocol support**

### 📊 Comprehensive Analysis
- **Recursive inventory processing** of nested group structures
- **Real-time validation** of inventory structure and projects
- **Detailed statistics** showing groups and project counts
- **Duplicate detection** and elimination
- **Group-based filtering** and targeted operations

### 🎨 Beautiful User Experience
- **Colorized output** with emojis and progress bars
- **Branch and description display** in `syncx list`
- **Verbose monitoring** with detailed progress tracking
- **Dry-run preview** for all operations
- **Clear error reporting** with actionable messages

---

## 🔧 Development Commands

### Using Make (Recommended)
```bash
make help       # Show all available commands
make build      # Build the application
make install    # Build and install globally
make uninstall  # Remove installation
make clean      # Clean build artifacts
make test       # Run tests
make fmt        # Format code
make run        # Run locally without installing
```

### Direct Commands
```bash
# Build the application
go build -o syncx main.go

# Run tests
go test ./...

# Format code
go fmt ./...

# Download dependencies
go mod download
```

---

## 🏗️ Architecture

This is a Go CLI application built with the Cobra framework. Architecture follows a clean separation of concerns:

- **`cmd/`** — Cobra commands; each file owns one command
- **`internal/`** — Core business logic, no CLI dependencies
  - `types.go` — Shared data structures (`Project`, `Group`, `Inventory`, `ProjectInfo`)
  - `git.go` — All git operations (clone with branch support, pull, status)
  - `inventory.go` — JSON loading, project collection, filtering
  - `tracker.go` — Smart state tracking between runs
  - `logger.go` — Colorized, emoji-rich output

### Dependencies
- `github.com/spf13/cobra` — CLI framework
- `github.com/fatih/color` — Terminal colors
- `github.com/schollz/progressbar/v3` — Progress bars
- `github.com/briandowns/spinner` — Loading spinners

---

## 🧪 Testing and Quality

- Run `go test ./...` for unit tests
- Use `go fmt ./...` to format code according to Go standards
- The application includes dry-run mode for safe operation testing
- Verbose logging available with `--verbose` flag for debugging

---

## 📄 License

See LICENSE file for details.

---

**Built with ❤️ using Go and Cobra**

*For additional development guidance and architectural details, see [CLAUDE.md](CLAUDE.md)*
