# Installation Guide - Olive Clone Assistant

## Quick Install

### Option 1: Using Make (Recommended)

```bash
# Clone the repository
git clone <your-repository-url>
cd olive-sync

# Install globally (one command does it all!)
make install
```

### Option 2: Using Install Script

```bash
# Clone the repository
git clone <your-repository-url>
cd olive-sync

# Run the installer
./scripts/install.sh
```

## What Happens During Installation?

The installer will:
1. ✅ Detect your OS and architecture (macOS/Linux, Intel/ARM)
2. ✅ Build the optimized binary for your platform
3. ✅ Install to `/usr/local/bin` (or `~/bin` if you don't have sudo)
4. ✅ Make the command available globally as `olive-clone`
5. ✅ Automatically configure your shell (bash/zsh) if needed

## Verify Installation

After installation, test that it works:

```bash
# Check if olive-clone is available
olive-clone --version

# Show help
olive-clone --help

# Try the interactive wizard
olive-clone wizard
```

## First Usage

Once installed, you can use `olive-clone` from anywhere:

```bash
# Interactive setup (recommended for first-time users)
olive-clone wizard

# Clone all repositories from inventory
olive-clone clone --file projects-inventory.json --protocol ssh -o ~/repositories

# Update existing repositories only
olive-clone pull --file projects-inventory.json -o ~/repositories

# List all available projects
olive-clone list --file projects-inventory.json --verbose

# Clone specific group
olive-clone clone --file projects-inventory.json --group "Team Ludus/Libraries" -o ~/repos
```

## Troubleshooting

### Command not found after installation

If you see "command not found" after installation:

1. **Restart your terminal** (most common fix)
2. Or manually reload your shell configuration:
   ```bash
   source ~/.bashrc    # for bash
   source ~/.zshrc     # for zsh
   ```

### Permission denied

If you get permission errors during installation:

- The installer will automatically ask for `sudo` when needed
- Alternatively, it will install to `~/bin` without requiring sudo

### Installation in custom location

If you want to install to a custom location:

```bash
# Build first
make build

# Copy to your preferred location
cp build/olive-clone /your/custom/path/

# Add to PATH
export PATH="/your/custom/path:$PATH"
```

## Uninstall

To completely remove the installation:

```bash
make uninstall
# OR
./scripts/uninstall.sh
```

This will:
- Remove the binary from your system
- Clean up shell configuration files
- Remove PATH entries added during installation

## Platform Support

✅ macOS (Intel & Apple Silicon)
✅ Linux (x86_64 & ARM64)
✅ Windows (via WSL or native with manual build)

## Requirements

- Go 1.21 or later (only for building from source)
- Git
- SSH keys configured (for SSH protocol)
- GitLab/GitHub access tokens (if using private repositories)

## Next Steps

After installation:

1. Read [CLAUDE.md](CLAUDE.md) for detailed command documentation
2. Check [README.md](README.md) for project overview
3. Run `olive-clone wizard` for interactive setup
4. Enjoy seamless repository management! 🫒
