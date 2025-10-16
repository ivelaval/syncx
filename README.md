# 🫒 Olive Clone Assistant v2.0

A modern, intelligent repository cloning tool built specifically for Olive.com projects with smart tracking and automatic directory management.

## 🚀 Quick Start

```bash
# Build the application
go build -o olive-clone main.go

# Clone all repositories with smart tracking
./olive-clone clone --file ../projects-inventory.json --protocol ssh -o /Users/vennet/Olive.com/uproarcar

# Interactive setup (recommended for first-time users)
./olive-clone wizard --file ../projects-inventory.json
```

## 📋 Commands Overview

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

## 🎯 Your Essential Commands (Ready to Use)

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

## ✨ Key Features

### 🎯 Smart Tracking System
- **Automatic directory creation** for missing group structures
- **Intelligent diff detection** between inventory JSON and physical structure
- **MD5-based change detection** for inventory updates
- **Persistent tracking** with `.olive-clone-tracker.json` files
- **Git change detection** using `git fetch` to check for remote updates

### 🚀 Enhanced Git Operations
- **Robust clone operations** with verification and error handling
- **Smart pull strategy** with fetch-first approach
- **Conflict resolution** with fallback mechanisms
- **Commit hash tracking** for change detection
- **SSH and HTTPS protocol support**

### 📊 Comprehensive Analysis
- **Recursive inventory processing** of nested group structures
- **Real-time validation** of inventory structure and projects
- **Detailed statistics** showing groups and project counts
- **Duplicate detection** and elimination
- **Group-based filtering** and targeted operations

### 🎨 Beautiful User Experience
- **Colorized output** with emojis and progress bars
- **Interactive wizard** for guided setup
- **Verbose monitoring** with detailed progress tracking
- **Dry-run preview** for all operations
- **Clear error reporting** with actionable messages

## 🔧 Development Commands

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

## 📁 Project Structure

```
olive-clone-assistant-v2/
├── main.go                 # Application entry point
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── README.md              # Complete usage documentation
├── CLAUDE.md              # Development guidance for Claude Code
│
├── cmd/                   # Cobra CLI commands
│   ├── root.go            # Root command and global flags
│   ├── clone.go           # Clone/update repositories with smart tracking
│   ├── pull.go            # Update existing repositories only
│   ├── list.go            # List projects and groups
│   ├── status.go          # Check repository status
│   ├── wizard.go          # Interactive wizard
│   └── processor.go       # Project processing with tracking
│
├── internal/              # Core functionality
│   ├── types.go           # Data structures and types
│   ├── logger.go          # Colored logging system
│   ├── git.go             # Git operations with enhanced tracking
│   ├── inventory.go       # Inventory file processing with smart analysis
│   ├── tracker.go         # Smart tracking system (NEW)
│   └── wizard.go          # Interactive wizard system
│
└── scripts/               # Build and utility scripts
    ├── build.sh           # Build script
    ├── install.sh         # Installation script
    └── release.sh         # Release preparation
```

## 🆕 Latest Improvements (v2.0)

### **🔄 Smart Tracking System**
- **Automatic project tracking** with `.olive-clone-tracker.json` files
- **Inventory change detection** using MD5 hashing
- **Smart diff analysis** to identify new, modified, and removed projects
- **Git change detection** to check for remote updates
- **Persistent state management** across multiple runs

### **📁 Intelligent Directory Management**
- **Automatic structure creation** based on URL mapping
- **Group hierarchy preservation** from JSON to physical directories
- **Conflict resolution** for existing directories
- **Path extraction** from GitLab URLs with proper nesting

### **🚀 Enhanced Performance**
- **Concurrent operations** with configurable parallelism
- **Smart skip logic** for up-to-date repositories
- **Optimized git operations** with fetch-first strategy
- **Progress tracking** with detailed operation statistics

### **🎯 Improved User Experience**
- **Clear operation modes** (clone, pull, wizard)
- **Group-based filtering** for targeted operations
- **Comprehensive dry-run preview** for all commands
- **Detailed verbose output** for monitoring and debugging
- **Interactive project selection** with wizard guidance

---

**🫒 Built with ❤️ for the Olive.com team**

*For additional development guidance and architectural details, see [CLAUDE.md](CLAUDE.md)*