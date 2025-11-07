#!/bin/bash

# Release preparation script for SyncX v2.0
set -e

VERSION=${1:-"v2.0.0"}
echo "🫒 Preparing SyncX ${VERSION} Release"
echo "=================================================="

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ Invalid version format. Use: v2.0.0"
    exit 1
fi

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf build/ release/

# Build all platforms
echo "🔨 Building all platforms..."
./scripts/build.sh "${VERSION#v}"

# Create release directory
mkdir -p release/

echo "📦 Creating release archives..."

# Create archives for each platform
for binary in build/syncx-*; do
    if [ -f "$binary" ]; then
        platform=$(basename "$binary" | sed 's/syncx-//')
        archive_name="syncx-assistant-${VERSION}-${platform}"
        
        # Create platform-specific directory
        mkdir -p "release/temp/${archive_name}"
        
        # Copy files
        cp "$binary" "release/temp/${archive_name}/syncx"
        cp README.md "release/temp/${archive_name}/"
        cp -r docs/ "release/temp/${archive_name}/"
        cp -r examples/ "release/temp/${archive_name}/"
        cp scripts/install.sh "release/temp/${archive_name}/"
        
        # Create archive
        cd release/temp/
        if [[ "$platform" == *"windows"* ]]; then
            zip -r "../${archive_name}.zip" "${archive_name}/"
        else
            tar -czf "../${archive_name}.tar.gz" "${archive_name}/"
        fi
        cd ../../
    fi
done

# Clean temporary files
rm -rf release/temp/

# Create checksums
echo "🔐 Creating checksums..."
cd release/
find . -name "*.tar.gz" -o -name "*.zip" | xargs shasum -a 256 > checksums.txt
cd ..

# Generate release notes
echo "📝 Generating release notes..."
cat > release/RELEASE_NOTES.md << EOF
# SyncX ${VERSION}

## 🎉 What's New

### GitCook-Inspired Interactive Experience
- **🧙‍♂️ Interactive Wizard** - Step-by-step guided repository management
- **🎯 Multi-Select Interface** - Choose exactly what to clone/update  
- **📊 Beautiful Progress Bars** - Visual feedback for all operations
- **⚡ Parallel Processing** - Configurable concurrent operations

### Enhanced Commands
- **\`wizard\`** - New interactive wizard command inspired by @vennet/gitcook
- **\`clone --interactive\`** - Enhanced with full wizard system
- **\`status\`** - Comprehensive repository health checking
- **\`list\`** - Beautiful inventory exploration

### Key Features
- Three wizard modes: Quick, Custom, and Advanced
- GitCook-style question flows with contextual guidance
- Preview and confirmation before operations
- Smart defaults with full customization options
- Cross-platform binaries for all major operating systems

## 📦 Installation

### Quick Install
\`\`\`bash
# Download and extract for your platform
tar -xzf syncx-assistant-${VERSION}-\$(uname -s | tr '[:upper:]' '[:lower:]')-\$(uname -m | sed 's/x86_64/amd64/').tar.gz
cd syncx-assistant-${VERSION}-*
./install.sh
\`\`\`

### Manual Install
1. Download the appropriate binary for your platform
2. Make it executable: \`chmod +x syncx\`
3. Move to your PATH: \`mv syncx /usr/local/bin/\`

## 🚀 Quick Start

\`\`\`bash
# Interactive wizard (Recommended)
syncx wizard

# Help and examples
syncx --help
syncx wizard --help
\`\`\`

## 📋 Platform Support

- macOS (Intel & Apple Silicon)
- Linux (x86_64 & ARM64)
- Windows (x86_64)

---

**Built with ❤️ for repository management excellence**
EOF

echo ""
echo "🎉 Release ${VERSION} prepared successfully!"
echo ""
echo "📂 Release artifacts:"
ls -la release/
echo ""
echo "🔐 Checksums:"
cat release/checksums.txt
echo ""
echo "📝 Release notes: release/RELEASE_NOTES.md"
echo "📦 Ready to publish!"