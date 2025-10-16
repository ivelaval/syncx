#!/bin/bash

# Version bump script
set -e

BUMP_TYPE=${1:-"patch"}
VERSION_FILE="VERSION"

# Check if VERSION file exists
if [ ! -f "$VERSION_FILE" ]; then
    echo "❌ VERSION file not found"
    exit 1
fi

# Read current version
CURRENT_VERSION=$(cat $VERSION_FILE | tr -d '\n')

# Parse version (assumes semantic versioning: MAJOR.MINOR.PATCH)
if [[ $CURRENT_VERSION =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    PATCH="${BASH_REMATCH[3]}"
else
    echo "❌ Invalid version format in VERSION file: $CURRENT_VERSION"
    echo "Expected format: MAJOR.MINOR.PATCH (e.g., 2.1.0)"
    exit 1
fi

# Bump version based on type
case $BUMP_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "❌ Invalid bump type: $BUMP_TYPE"
        echo "Valid types: major, minor, patch"
        exit 1
        ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "📊 Current version: $CURRENT_VERSION"
echo "📈 New version:     $NEW_VERSION"
echo ""

# Confirm
read -p "Proceed with version bump? (y/N): " -r
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Version bump cancelled"
    exit 0
fi

# Update VERSION file
echo "$NEW_VERSION" > $VERSION_FILE

echo "✅ Version bumped to $NEW_VERSION"
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff VERSION"
echo "  2. Commit: git commit -am 'Bump version to v${NEW_VERSION}'"
echo "  3. Tag: git tag -a v${NEW_VERSION} -m 'Release v${NEW_VERSION}'"
echo "  4. Build: make build"
echo "  5. Push: git push && git push --tags"
