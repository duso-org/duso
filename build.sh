#!/bin/bash
set -e

# Embed stdlib, docs, and contrib directories
go generate ./cmd/duso

# Get version from git tags
VERSION=$(git describe --tags 2>/dev/null || echo "dev")

# Create bin directory if it doesn't exist
mkdir -p bin

# Build with version embedded
go build -ldflags "-X main.Version=$VERSION" -o bin/duso ./cmd/duso

echo "Built bin/duso $VERSION"
