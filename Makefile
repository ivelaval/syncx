.PHONY: help build build-dev install install-dev uninstall clean test fmt run version bump-patch bump-minor bump-major

# Get version from VERSION file
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")

# Default target
help:
	@echo "🫒 Olive Clone Assistant v$(VERSION) - Available Commands"
	@echo "=========================================================="
	@echo ""
	@echo "Development:"
	@echo "  make build-dev     - Quick build for development (no cross-compile)"
	@echo "  make install-dev   - Build and install for development"
	@echo "  make run           - Run locally without installing"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format code"
	@echo ""
	@echo "Production:"
	@echo "  make build         - Build all platforms with version"
	@echo "  make install       - Build and install globally"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Versioning:"
	@echo "  make version       - Show current version"
	@echo "  make bump-patch    - Increment patch version (2.1.0 -> 2.1.1)"
	@echo "  make bump-minor    - Increment minor version (2.1.0 -> 2.2.0)"
	@echo "  make bump-major    - Increment major version (2.1.0 -> 3.0.0)"
	@echo ""
	@echo "Other:"
	@echo "  make uninstall     - Remove installation"
	@echo ""
	@echo "Quick Development Workflow:"
	@echo "  1. Edit code"
	@echo "  2. make install-dev    # Fast install for testing"
	@echo "  3. olive-clone --version"
	@echo ""
	@echo "Release Workflow:"
	@echo "  1. make bump-minor     # Update version"
	@echo "  2. git commit -am 'Bump version to vX.Y.Z'"
	@echo "  3. make build          # Build all platforms"
	@echo "  4. make install        # Install locally"
	@echo ""

# Show current version
version:
	@echo "Current version: $(VERSION)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Git branch: $$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"

# Quick build for development (current platform only)
build-dev:
	@echo "🔨 Building olive-clone for development..."
	@VERSION=$$(cat VERSION) && \
	BUILD_TIME=$$(date -u '+%Y-%m-%d_%H:%M:%S') && \
	GIT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "dev") && \
	LDFLAGS="-X olive-clone-assistant-v2/cmd.Version=$$VERSION-dev -X olive-clone-assistant-v2/cmd.BuildTime=$$BUILD_TIME -X olive-clone-assistant-v2/cmd.GitCommit=$$GIT_COMMIT" && \
	go build -ldflags "$$LDFLAGS" -o olive-clone main.go
	@echo "✅ Development build complete: ./olive-clone"

# Build all platforms (production)
build:
	@echo "🔨 Building olive-clone for all platforms..."
	@./scripts/build.sh

# Quick install for development
install-dev: build-dev
	@echo "📦 Installing development build..."
	@mkdir -p ~/bin
	@cp olive-clone ~/bin/olive-clone
	@chmod +x ~/bin/olive-clone
	@echo "✅ Development version installed to ~/bin/olive-clone"
	@echo ""
	@echo "🎉 Test with: olive-clone --version"

# Build and install (production)
install: build
	@./scripts/install.sh

# Uninstall
uninstall:
	@./scripts/uninstall.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf build/
	@rm -f olive-clone olive-clone-*
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./... -v

# Format code
fmt:
	@echo "✨ Formatting code..."
	@go fmt ./...
	@echo "✅ Format complete"

# Run locally (without installing)
run:
	@go run main.go

# Bump version numbers
bump-patch:
	@echo "📈 Bumping patch version..."
	@./scripts/bump-version.sh patch

bump-minor:
	@echo "📈 Bumping minor version..."
	@./scripts/bump-version.sh minor

bump-major:
	@echo "📈 Bumping major version..."
	@./scripts/bump-version.sh major
