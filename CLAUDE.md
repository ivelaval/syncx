# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Installation

### Quick Install (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd olive-sync

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

## Development Commands

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

# Build for development with scripts
./scripts/build.sh

# Install system-wide
./scripts/install.sh

# Run tests
go test ./...

# Format code
go fmt ./...

# Download dependencies
go mod download
```

## Commands Overview

### Available Commands
| Command | Purpose | Best For |
|---------|---------|----------|
| `clone` | Clone new + update existing | Daily sync, full repository management |
| `pull` | Update existing projects only | Quick updates without new clones |
| `check` | Check for uncommitted local changes | Pre-sync validation, change detection |
| `scan` | Recursively scan directory for git repos with changes | No inventory needed, workspace scanning |
| `list` | Show projects and groups | Discovery, validation |
| `status` | Check repository status | Monitoring, troubleshooting |

### Operation Modes Comparison
| Feature | `clone` | `pull` |
|---------|---------|--------|
| Clone new projects | ✅ | ❌ |
| Update existing | ✅ | ✅ |
| Smart tracking | ✅ | ✅ |
| Group filtering | ✅ | ✅ |
| Parallel processing | ✅ | ✅ |

### Advanced Usage Tips
| Scenario | Recommended Command |
|----------|-------------------|
| **Clone only new projects** | Use fresh output directory: `-o /new/path` |
| **Clone specific new project** | Filter by group: `--group "NewGroupName"` |
| **Preview before action** | Add `--dry-run -v` to any command |
| **Initial environment setup** | `clone` to fresh directory |
| **Daily sync workflow** | `clone` (handles both new and updates) |
| **Update only existing** | `pull` (safe for preserving local changes) |

### Application Commands

## 🚀 Quick One-Line Commands (Most Used)

### Clone & Update Everything (Smart Mode)
```bash
# Clone new projects + update existing ones (recommended)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar

# Same but with verbose output
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v

# Preview what will happen (dry run)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --dry-run -v
```

### Clone Only New Projects (Skip Updates)
```bash
# Method 1: Use dry-run to preview, then manually select
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --dry-run -v
# Review output and run again without --dry-run if needed

# Method 2: Clone to a fresh directory (guarantees only new clones)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new

# Method 3: Filter by specific new groups 
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "NewGroup"
```

### Update Only Existing Projects
```bash
# Update existing projects only (no new clones)
./syncx pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar

# Update with verbose output
./syncx pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar -v
```

### Target Specific Groups
```bash
# Clone/update specific group
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Tools/Sales"

# Clone only new projects in specific group (use fresh directory)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new --group "Team Ludus/Libraries"

# Update only existing projects in specific group
./syncx pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar --group "Analytics"
```

## 📋 Exploration & Discovery Commands

### Project Discovery
```bash
# List all projects and groups
./syncx list --file ../projects-inventory.json --verbose

# Show only available groups
./syncx clone --file ../projects-inventory.json --show-groups

# Show groups for any command
./syncx pull --file ../projects-inventory.json --show-groups

# Check status of existing repositories
./syncx status --output /Users/vennet/Olive.com/uproarcar --verbose
```

## ⚙️ Advanced Configuration Commands

### Protocol Options
```bash
# Use SSH (default, recommended)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /path

# Use HTTPS (for environments without SSH keys)
./syncx clone --file ../projects-inventory.json --protocol http -o /path
```

### Parallel Processing
```bash
# Process multiple repositories in parallel (faster)
./syncx clone --file ../projects-inventory.json -o /path --parallel 5

# Pull with parallel processing
./syncx pull --file ../projects-inventory.json -o /path --parallel 3
```

## 📊 Monitoring & Validation Commands

### Dry Run (Preview)
```bash
# Preview all operations
./syncx clone --file ../projects-inventory.json -o /path --dry-run -v

# Preview clone to fresh directory (guarantees only new clones)
./syncx clone --file ../projects-inventory.json -o /fresh/path --dry-run -v

# Preview pull operations
./syncx pull --file ../projects-inventory.json -o /path --dry-run -v
```

### Status & Validation
```bash
# Check for uncommitted local changes in all repositories
./syncx check --file ../projects-inventory.json -o /path

# Check with verbose output
./syncx check --file ../projects-inventory.json -o /path -v

# Check specific group for uncommitted changes
./syncx check --file ../projects-inventory.json -o /path --group "Analytics"

# Check what needs updating
./syncx status --file ../projects-inventory.json -o /path -v

# Validate inventory file
./syncx list --file ../projects-inventory.json

# Show detailed statistics
./syncx clone --file ../projects-inventory.json --show-groups
```

## 🎯 Use Case Examples

### Initial Environment Setup
```bash
# Clone all repositories to set up environment
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v
```

### Daily Development Workflow
```bash
# First, check for uncommitted changes before syncing
./syncx check --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar -v

# Update existing projects only (preserve local changes in new dirs)
./syncx pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar -v

# Or full sync (clone new + update existing)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v
```

### Adding New Team Projects
```bash
# Clone only new projects from recent inventory updates (use fresh directory)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new -v
```

### Working with Specific Teams
```bash
# Get all Team Ludus projects
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Team Ludus/Libraries" -v

# Update only Team Sharks projects
./syncx pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar --group "Team Sharks/Microservices" -v
```

### Managing Multiple Repository Collections
```bash
# Scenario: You have multiple clones for different purposes

# Production environment clone
./syncx clone --file ../projects-inventory.json --protocol ssh -o ~/production-repos
./syncx check --file ../projects-inventory.json -o ~/production-repos

# Development environment clone
./syncx clone --file ../projects-inventory.json --protocol ssh -o ~/development-repos
./syncx check --file ../projects-inventory.json -o ~/development-repos

# Backup/archive clone
./syncx pull --file ../projects-inventory.json -o ~/backup-repos
./syncx check --file ../projects-inventory.json -o ~/backup-repos

# Personal workspace
./syncx clone --file ../projects-inventory.json --protocol ssh -o ~/workspace/olive
./syncx check --file ../projects-inventory.json -o ~/workspace/olive

# Check all collections for uncommitted changes
./syncx check --file ../projects-inventory.json -o ~/production-repos -v
./syncx check --file ../projects-inventory.json -o ~/development-repos -v
./syncx check --file ../projects-inventory.json -o ~/backup-repos -v
```

## 🎯 Your Requested Commands (Ready to Use)

### Complete Repository Sync
```bash
# Clone all projects + update existing ones (recommended daily command)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar

# Same with verbose output for monitoring
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v

# Preview what would happen before executing
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --dry-run -v
```

### Clone Only New Projects (No Updates)
```bash
# Clone only new projects to fresh directory (guarantees no updates)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new

# Preview clone-only to fresh directory
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new --dry-run -v
```

### Update Only Existing Projects
```bash
# Update existing projects only (no new clones)
./syncx pull --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar

# Update with verbose monitoring
./syncx pull --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v
```

### Check for Uncommitted Changes (with inventory)
```bash
# Check all repositories for uncommitted changes
./syncx check --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar

# Check with verbose output to see clean repositories too
./syncx check --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar -v

# Check specific group for uncommitted changes
./syncx check --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar --group "Analytics"

# Check with parallel processing for faster results
./syncx check --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar --parallel 20

# Check a different clone location (useful if you have multiple clones)
./syncx check --file ../projects-inventory.json -o /Users/vennet/Olive.com/backup-repos

# Check personal workspace clone
./syncx check --file ../projects-inventory.json -o ~/workspace/olive-projects

# Check production clone vs development clone
./syncx check --file ../projects-inventory.json -o /Users/vennet/production-repos
./syncx check --file ../projects-inventory.json -o /Users/vennet/development-repos
```

### Scan for Changes Recursively (NO inventory needed!)
```bash
# Scan current directory for git repositories with changes
./syncx scan .

# Scan specific directory
./syncx scan /Users/vennet/Olive.com/uproarcar

# Scan with verbose output (shows full paths)
./syncx scan ~/workspace -v

# Scan with limited depth (faster for large directories)
./syncx scan /Users/vennet/Olive.com -d 3

# Scan and show clean repositories too
./syncx scan . --show-clean

# Scan home directory for all git repos with changes
./syncx scan ~ -d 5

# Scan with more parallel processing for speed
./syncx scan /Users/vennet/Olive.com --parallel 20 -d 4

# Scan multiple locations
./syncx scan ~/production-repos
./syncx scan ~/dev-repos
./syncx scan ~/workspace

# Quick scan of common locations
for dir in ~/workspace ~/projects ~/Documents; do
  echo "Scanning $dir..."
  ./syncx scan "$dir" -d 3
done
```

### Group-Specific Operations
```bash
# Target specific group (clone + update)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Tools/Sales"

# Update only specific group
./syncx pull --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Analytics"

# Clone specific group to fresh location (new projects only)
./syncx clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-tools --group "Tools"
```

## Architecture

This is a Go CLI application built with the Cobra framework for repository management at Olive.com. The architecture follows a clean separation of concerns:

### Core Structure
- **`main.go`** - Application entry point that delegates to cmd package
- **`cmd/`** - Cobra CLI commands and command-line interface logic
  - `root.go` - Root command with global flags and configuration
  - `clone.go` - Repository cloning and updating functionality
  - `pull.go` - Pull updates for existing repositories
  - `list.go` - Project inventory listing and exploration
  - `status.go` - Repository health checking
- **`internal/`** - Core business logic and data structures
  - `types.go` - Data structures (Project, Group, Inventory, OperationResult)
  - `git.go` - Git operations (clone, pull, status checking)
  - `inventory.go` - JSON inventory file processing
  - `logger.go` - Colored logging and output formatting
  - `tracker.go` - Smart repository tracking system

### Key Concepts
- **Inventory System**: Projects are organized in JSON files with hierarchical groups
- **Protocol Support**: Both SSH and HTTP git protocols
- **Smart Tracking System**: Tracks repository state and optimizes operations
- **Parallel Processing**: Configurable concurrent operations for performance
- **Directory Structure**: All projects are organized under a `projects/` subdirectory within the base directory, with cleaned paths that remove redundant organizational prefixes (e.g., `uproarcar/olive-com`)

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Terminal colors
- `github.com/schollz/progressbar/v3` - Progress bars
- `github.com/briandowns/spinner` - Loading spinners

## Directory Structure

When you run `syncx clone` with a base directory (e.g., `/Users/vennet/Olive.com`), all projects are organized under a `projects/` subdirectory with a cleaned structure:

```
/Users/vennet/Olive.com/
├── projects/                          # All repositories go here
│   ├── analytics/
│   │   ├── fenske/
│   │   └── pyspark-test/
│   ├── salesforce/
│   │   ├── olive/
│   │   ├── ford/
│   │   └── sf-data-analysis/
│   ├── artificial-intelligence/
│   │   ├── mcp-development/
│   │   └── bedrock-development-agent/
│   ├── devops/
│   └── tools/
│       └── libraries/
│           └── OEM-SF/
└── .olive-clone-tracker.json         # Tracking file (auto-generated)
```

### Path Cleaning Logic

The application automatically cleans Git URLs by removing redundant organizational prefixes:

- **Original URL**: `gitlab.com:uproarcar/olive-com/analytics/fenske.git`
- **Cleaned Path**: `/Users/vennet/Olive.com/projects/analytics/fenske`

Prefixes automatically removed:
- `uproarcar`
- `olive-com`
- `olive.com`
- `olivecom`

This keeps your directory structure clean and focused on the actual project categories rather than organizational structure.

## Project Inventory Format

The application expects a `projects-inventory.json` file with this structure:
```json
{
  "phisical-location": "optional-location",
  "groups": [
    {
      "name": "Frontend",
      "projects": [
        {"name": "project-name", "url": "git-url"}
      ],
      "groups": []
    }
  ],
  "projects": []
}
```

## Testing and Quality

- Run `go test ./...` for unit tests
- Use `go fmt ./...` to format code according to Go standards
- The application includes dry-run mode for safe operation testing
- Verbose logging available with `--verbose` flag for debugging
- Add to memorize
- Add to memorize