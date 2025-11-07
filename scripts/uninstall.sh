#!/bin/bash

# Uninstallation script for SyncX
set -e

echo "🫒 Uninstalling SyncX"
echo "==========================================="
echo ""

# Find installation location
if command -v syncx &> /dev/null; then
    INSTALL_PATH=$(which syncx)
    INSTALL_DIR=$(dirname "$INSTALL_PATH")

    echo "📍 Found installation: $INSTALL_PATH"
    echo ""

    read -p "Are you sure you want to uninstall? (y/N): " -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Uninstall cancelled"
        exit 0
    fi

    # Remove binary
    echo "🗑️  Removing binary from $INSTALL_DIR..."
    if [ -w "$INSTALL_DIR" ]; then
        rm -f "$INSTALL_PATH"
    else
        echo "🔒 Administrator privileges required"
        sudo rm -f "$INSTALL_PATH"
    fi

    # Clean up PATH entries in shell RC files
    echo "🧹 Cleaning up shell configuration files..."

    for rc_file in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc"; do
        if [ -f "$rc_file" ]; then
            if grep -q "# Added by syncx installer" "$rc_file"; then
                echo "   Cleaning $rc_file..."
                # Remove the comment line and the next line (PATH export)
                sed -i.bak '/# Added by syncx installer/,+1d' "$rc_file"
                rm -f "${rc_file}.bak"
            fi
        fi
    done

    echo ""
    echo "✅ Uninstallation complete!"
    echo ""
    echo "💡 Note: You may need to restart your terminal or run:"
    echo "   source ~/.bashrc  # or ~/.zshrc"

else
    echo "❌ syncx is not installed or not in PATH"
    echo ""
    echo "🔍 Searching for installations in common locations..."

    FOUND=false
    for dir in "/usr/local/bin" "$HOME/bin" "$HOME/.local/bin"; do
        if [ -f "$dir/syncx" ]; then
            echo "   Found: $dir/syncx"
            FOUND=true
        fi
    done

    if [ "$FOUND" = false ]; then
        echo "   No installations found"
    else
        echo ""
        echo "💡 Manually remove files if needed"
    fi

    exit 1
fi
