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

After installation, the `olive-clone` command will be available globally:
```bash
olive-clone --help
olive-clone wizard
olive-clone clone --file projects-inventory.json --protocol ssh -o ~/repos
```

### Manual Installation

```bash
# Build the binary
go build -o olive-clone main.go

# Copy to your PATH (choose one)
sudo cp olive-clone /usr/local/bin/
# OR
cp olive-clone ~/bin/  # Make sure ~/bin is in your PATH
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
go build -o olive-clone main.go

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
| `wizard` | Interactive guided setup | First-time users, complex configurations |
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
| Interactive mode | ✅ | ✅ |

### Advanced Usage Tips
| Scenario | Recommended Command |
|----------|-------------------|
| **Clone only new projects** | Use fresh output directory: `-o /new/path` |
| **Clone specific new project** | Filter by group: `--group "NewGroupName"` |
| **Preview before action** | Add `--dry-run -v` to any command |
| **Initial environment setup** | `wizard` or `clone` to fresh directory |
| **Daily sync workflow** | `clone` (handles both new and updates) |
| **Update only existing** | `pull` (safe for preserving local changes) |

### Application Commands

## 🚀 Quick One-Line Commands (Most Used)

### Clone & Update Everything (Smart Mode)
```bash
# Clone new projects + update existing ones (recommended)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar

# Same but with verbose output
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v

# Preview what will happen (dry run)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --dry-run -v
```

### Clone Only New Projects (Skip Updates)
```bash
# Method 1: Use dry-run to preview, then manually select
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --dry-run -v
# Review output and run again without --dry-run if needed

# Method 2: Clone to a fresh directory (guarantees only new clones)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new

# Method 3: Filter by specific new groups 
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "NewGroup"
```

### Update Only Existing Projects
```bash
# Update existing projects only (no new clones)
./olive-clone pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar

# Update with verbose output
./olive-clone pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar -v
```

### Target Specific Groups
```bash
# Clone/update specific group
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Tools/Sales"

# Clone only new projects in specific group (use fresh directory)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new --group "Team Ludus/Libraries"

# Update only existing projects in specific group
./olive-clone pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar --group "Analytics"
```

## 🧙‍♂️ Interactive & Discovery Commands

### Wizard (Guided Setup)
```bash
# Interactive wizard (best for first-time users)
./olive-clone wizard --file ../projects-inventory.json

# Wizard with specific output directory
./olive-clone wizard --file ../projects-inventory.json -o /custom/path
```

### Exploration & Discovery
```bash
# List all projects and groups
./olive-clone list --file ../projects-inventory.json --verbose

# Show only available groups
./olive-clone clone --file ../projects-inventory.json --show-groups

# Show groups for any command
./olive-clone pull --file ../projects-inventory.json --show-groups

# Check status of existing repositories
./olive-clone status --output /Users/vennet/Olive.com/uproarcar --verbose
```

## ⚙️ Advanced Configuration Commands

### Protocol Options
```bash
# Use SSH (default, recommended)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /path

# Use HTTPS (for environments without SSH keys)
./olive-clone clone --file ../projects-inventory.json --protocol http -o /path
```

### Parallel Processing
```bash
# Process multiple repositories in parallel (faster)
./olive-clone clone --file ../projects-inventory.json -o /path --parallel 5

# Pull with parallel processing
./olive-clone pull --file ../projects-inventory.json -o /path --parallel 3
```

### Interactive Mode
```bash
# Interactive project selection within any command
./olive-clone clone --file ../projects-inventory.json -o /path --interactive

# Interactive pull
./olive-clone pull --file ../projects-inventory.json -o /path --interactive
```

## 📊 Monitoring & Validation Commands

### Dry Run (Preview)
```bash
# Preview all operations
./olive-clone clone --file ../projects-inventory.json -o /path --dry-run -v

# Preview clone to fresh directory (guarantees only new clones)
./olive-clone clone --file ../projects-inventory.json -o /fresh/path --dry-run -v

# Preview pull operations
./olive-clone pull --file ../projects-inventory.json -o /path --dry-run -v
```

### Status & Validation
```bash
# Check what needs updating
./olive-clone status --file ../projects-inventory.json -o /path -v

# Validate inventory file
./olive-clone list --file ../projects-inventory.json

# Show detailed statistics
./olive-clone clone --file ../projects-inventory.json --show-groups
```

## 🎯 Use Case Examples

### Initial Environment Setup
```bash
# 1. First time setup - use wizard
./olive-clone wizard --file ../projects-inventory.json

# 2. Or direct clone everything
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v
```

### Daily Development Workflow
```bash
# Update existing projects only (preserve local changes in new dirs)
./olive-clone pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar -v

# Or full sync (clone new + update existing)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v
```

### Adding New Team Projects
```bash
# Clone only new projects from recent inventory updates (use fresh directory)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new -v
```

### Working with Specific Teams
```bash
# Get all Team Ludus projects
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Team Ludus/Libraries" -v

# Update only Team Sharks projects
./olive-clone pull --file ../projects-inventory.json -o /Users/vennet/Olive.com/uproarcar --group "Team Sharks/Microservices" -v
```

## 🎯 Your Requested Commands (Ready to Use)

### Complete Repository Sync
```bash
# Clone all projects + update existing ones (recommended daily command)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar

# Same with verbose output for monitoring
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v

# Preview what would happen before executing
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --dry-run -v
```

### Clone Only New Projects (No Updates)
```bash
# Clone only new projects to fresh directory (guarantees no updates)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new

# Preview clone-only to fresh directory
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-new --dry-run -v
```

### Update Only Existing Projects
```bash
# Update existing projects only (no new clones)
./olive-clone pull --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar

# Update with verbose monitoring
./olive-clone pull --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar -v
```

### Group-Specific Operations
```bash
# Target specific group (clone + update)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Tools/Sales"

# Update only specific group
./olive-clone pull --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar --group "Analytics"

# Clone specific group to fresh location (new projects only)
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar-tools --group "Tools"
```

## Architecture

This is a Go CLI application built with the Cobra framework for repository management at Olive.com. The architecture follows a clean separation of concerns:

### Core Structure
- **`main.go`** - Application entry point that delegates to cmd package
- **`cmd/`** - Cobra CLI commands and command-line interface logic
  - `root.go` - Root command with global flags and configuration
  - `clone.go` - Repository cloning and updating functionality
  - `list.go` - Project inventory listing and exploration
  - `status.go` - Repository health checking
  - `wizard.go` - Interactive guided workflow
- **`internal/`** - Core business logic and data structures
  - `types.go` - Data structures (Project, Group, Inventory, OperationResult)
  - `git.go` - Git operations (clone, pull, status checking)
  - `inventory.go` - JSON inventory file processing
  - `logger.go` - Colored logging and output formatting
  - `wizard.go` - Interactive wizard system with step navigation

### Key Concepts
- **Inventory System**: Projects are organized in JSON files with hierarchical groups
- **Protocol Support**: Both SSH and HTTP git protocols
- **Smart Output Directory**: Defaults to `../repositories` (outside script folder)
- **Parallel Processing**: Configurable concurrent operations for performance
- **Interactive Modes**: Multi-step wizards with preview and confirmation

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Terminal colors
- `github.com/manifoldco/promptui` - Interactive prompts
- `github.com/schollz/progressbar/v3` - Progress bars
- `github.com/eiannone/keyboard` - Keyboard input handling

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