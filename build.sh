#!/bin/bash
set -e

# Embed stdlib, docs, and contrib directories
go generate ./cmd/duso

# Get version from git tags
VERSION=$(git describe --tags 2>/dev/null || echo "dev")

# Build with version embedded
go build -ldflags "-X main.Version=$VERSION" -o duso ./cmd/duso

echo "Built duso $VERSION"
