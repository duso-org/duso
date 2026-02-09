#!/bin/bash
set -e

# Ensure we're in the project root
if [ ! -f "go.mod" ] || [ ! -d "cmd/duso" ]; then
  echo "Error: build.sh must be run from project root"
  exit 1
fi

# Embed stdlib, docs, and contrib directories
go generate ./cmd/duso

# Get version from git tags
VERSION=$(git describe --tags 2>/dev/null || echo "dev")

# Create bin directory if it doesn't exist
mkdir -p bin

# Build with version embedded
go build -ldflags "-s -w -X main.Version=$VERSION" -trimpath -o bin/duso ./cmd/duso

# Clean up embedded file copies (they're in .gitignore and will be regenerated on next build)
rm -rf cmd/duso/stdlib cmd/duso/docs cmd/duso/contrib cmd/duso/examples
rm -f cmd/duso/README.md cmd/duso/CONTRIBUTING.md cmd/duso/LICENSE

echo "Built bin/duso $VERSION"
