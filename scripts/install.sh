#!/bin/bash

# Installation script for SyncX
set -e

echo "⚡ Installing SyncX"
echo "===================="
echo ""

# Check if binary exists
if [ ! -f "build/syncx" ]; then
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

BINARY_NAME="syncx"
if [ -f "build/syncx-${OS}-${ARCH}" ]; then
    BINARY_NAME="syncx-${OS}-${ARCH}"
fi

echo "🔍 Detected platform: ${OS}/${ARCH}"
echo "📦 Using binary: build/${BINARY_NAME}"
echo ""

# Check for installation directory preference
INSTALL_DIR=""
NEEDS_SUDO=false

if command -v syncx &> /dev/null; then
    CURRENT_PATH=$(which syncx)
    echo "⚠️  Existing installation found: $CURRENT_PATH"
    read -p "Replace existing installation? (y/N): " -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        INSTALL_DIR=$(dirname "$CURRENT_PATH")
        [ ! -w "$INSTALL_DIR" ] && NEEDS_SUDO=true
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
    sudo cp "build/${BINARY_NAME}" "$INSTALL_DIR/syncx"
    sudo chmod +x "$INSTALL_DIR/syncx"
else
    cp "build/${BINARY_NAME}" "$INSTALL_DIR/syncx"
    chmod +x "$INSTALL_DIR/syncx"
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
        echo '# Added by syncx installer' >> "$SHELL_RC"
        echo 'export PATH="$HOME/bin:$PATH"' >> "$SHELL_RC"
        echo "✅ Added to $SHELL_RC"
        echo "⚠️  Please run: source $SHELL_RC"
    fi
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "🎉 Quick Start:"
echo "   syncx --help                    # Show all commands"
echo "   syncx wizard                    # Interactive setup"
echo "   syncx clone --file inventory.json --protocol ssh -o ~/repos"
echo ""

# Test installation
echo "🔍 Testing installation..."
if command -v syncx &> /dev/null; then
    echo "✅ syncx is available globally"
    syncx --version 2>/dev/null || echo "Version: 2.0.0"
else
    echo "⚠️  syncx is not yet in PATH"
    echo "   Run: export PATH=\"$INSTALL_DIR:\$PATH\""
    echo "   Or restart your terminal"
fi

echo ""
echo "📚 For uninstall, run: ./scripts/uninstall.sh"