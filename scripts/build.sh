#!/bin/bash

# Build script for Olive Clone Assistant
set -e

# Get version from VERSION file or use provided argument
if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION | tr -d '\n')
else
    VERSION=${1:-"dev"}
fi

# Allow override via argument
if [ -n "$1" ]; then
    VERSION=$1
fi

echo "⚡ Building SyncX v${VERSION}"
echo "=========================================="

# Get build information
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Build flags (inject into cmd package)
LDFLAGS="-X olive-clone-assistant-v2/cmd.Version=${VERSION} -X olive-clone-assistant-v2/cmd.BuildTime=${BUILD_TIME} -X olive-clone-assistant-v2/cmd.GitCommit=${GIT_COMMIT}"

# Create build directory
mkdir -p build/

echo "📦 Building for current platform..."
go build -ldflags "${LDFLAGS}" -o build/syncx main.go

echo "🔨 Building cross-platform binaries..."

# Build for different platforms
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

for platform in "${PLATFORMS[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    output_name="syncx-${GOOS}-${GOARCH}"
    if [ $GOOS = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    echo "Building for ${GOOS}/${GOARCH}..."
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "${LDFLAGS}" \
        -o build/$output_name main.go
done

echo ""
echo "✅ Build complete!"
echo "📂 Binaries available in ./build/"
echo ""
ls -la build/
echo ""
echo "🚀 To install locally, run: ./scripts/install.sh"