#!/bin/bash

# Installation script for Olive Clone Assistant v2.0
set -e

echo "🫒 Installing Olive Clone Assistant v2.0"
echo "========================================"
echo ""

# Check if binary exists
if [ ! -f "build/olive-clone" ]; then
    echo "❌ Binary not found. Running build first..."
    ./scripts/build.sh
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64) ARCH="arm64" ;;
    aarch64) ARCH="arm64" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY_NAME="olive-clone"
if [ -f "build/olive-clone-${OS}-${ARCH}" ]; then
    BINARY_NAME="olive-clone-${OS}-${ARCH}"
fi

echo "🔍 Detected platform: ${OS}/${ARCH}"
echo "📦 Using binary: build/${BINARY_NAME}"
echo ""

# Check for installation directory preference
INSTALL_DIR=""
NEEDS_SUDO=false

if command -v olive-clone &> /dev/null; then
    CURRENT_PATH=$(which olive-clone)
    echo "⚠️  Existing installation found: $CURRENT_PATH"
    read -p "Replace existing installation? (y/N): " -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        INSTALL_DIR=$(dirname "$CURRENT_PATH")
    else
        echo "❌ Installation cancelled"
        exit 1
    fi
fi

# Determine installation directory
if [ -z "$INSTALL_DIR" ]; then
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    elif [ -d "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
        NEEDS_SUDO=true
    elif [ -w "$HOME/bin" ]; then
        INSTALL_DIR="$HOME/bin"
    else
        # Create ~/bin if it doesn't exist
        mkdir -p "$HOME/bin"
        INSTALL_DIR="$HOME/bin"
    fi
fi

echo "📂 Installing to: $INSTALL_DIR"

# Copy binary
if [ "$NEEDS_SUDO" = true ]; then
    echo "🔒 Administrator privileges required for installation to $INSTALL_DIR"
    sudo cp "build/${BINARY_NAME}" "$INSTALL_DIR/olive-clone"
    sudo chmod +x "$INSTALL_DIR/olive-clone"
else
    cp "build/${BINARY_NAME}" "$INSTALL_DIR/olive-clone"
    chmod +x "$INSTALL_DIR/olive-clone"
fi

# Add to PATH if needed
if [ "$INSTALL_DIR" = "$HOME/bin" ] && [[ ":$PATH:" != *":$HOME/bin:"* ]]; then
    echo ""
    echo "📝 Adding $HOME/bin to PATH..."

    # Detect shell and add to appropriate RC file
    if [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
        [ -f "$HOME/.bash_profile" ] && SHELL_RC="$HOME/.bash_profile"
    elif [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    else
        # Default to bashrc
        SHELL_RC="$HOME/.bashrc"
    fi

    # Add to shell config if not already present
    if [ -f "$SHELL_RC" ] && ! grep -q 'export PATH="$HOME/bin:$PATH"' "$SHELL_RC"; then
        echo '' >> "$SHELL_RC"
        echo '# Added by olive-clone installer' >> "$SHELL_RC"
        echo 'export PATH="$HOME/bin:$PATH"' >> "$SHELL_RC"
        echo "✅ Added to $SHELL_RC"
        echo "⚠️  Please run: source $SHELL_RC"
    fi
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "🎉 Quick Start:"
echo "   olive-clone --help                    # Show all commands"
echo "   olive-clone wizard                    # Interactive setup"
echo "   olive-clone clone --file inventory.json --protocol ssh -o ~/repos"
echo ""

# Test installation
echo "🔍 Testing installation..."
if command -v olive-clone &> /dev/null; then
    echo "✅ olive-clone is available globally"
    olive-clone --version 2>/dev/null || echo "Version: 2.0.0"
else
    echo "⚠️  olive-clone is not yet in PATH"
    echo "   Run: export PATH=\"$INSTALL_DIR:\$PATH\""
    echo "   Or restart your terminal"
fi

echo ""
echo "📚 For uninstall, run: ./scripts/uninstall.sh"