#!/bin/bash
set -e

# Ensure we're in the project root
if [ ! -f "go.mod" ] || [ ! -d "cmd/duso" ]; then
  echo "Error: build.sh must be run from project root"
  exit 1
fi

# Configuration
VERSION=$(git describe --tags 2>/dev/null || echo "dev")

# Platform/architecture combinations to build
# Format: "GOOS/GOARCH:binary_name"
PLATFORMS=(
  "darwin/amd64:duso-macos-intel:duso"
  "darwin/arm64:duso-macos-silicon:duso"
  "linux/amd64:duso-linux-amd64:duso"
  "linux/arm64:duso-linux-arm64:duso"
  "windows/amd64:duso-windows-amd64:duso.exe"
)

# Parse command line arguments
BUILD_MACOS=false
BUILD_LINUX=false
BUILD_WINDOWS=false
BUILD_ALL=false
CLEAN_ONLY=false

print_usage() {
  cat << EOF
Usage: ./build.sh [OPTIONS]

Build Duso binaries for different platforms and architectures.

OPTIONS:
  (no args)       Build for current platform only
  --all           Build for all platforms (macOS, Linux, Windows)
  --macos         Build for macOS (Intel + Apple Silicon)
  --linux         Build for Linux (amd64 + arm64)
  --windows       Build for Windows (amd64)
  --clean         Remove all built binaries
  --help          Show this message

EXAMPLES:
  ./build.sh                # Build for current platform
  ./build.sh --all          # Build for all platforms
  ./build.sh --macos --linux # Build for macOS and Linux only
EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --all)
      BUILD_ALL=true
      shift
      ;;
    --macos)
      BUILD_MACOS=true
      shift
      ;;
    --linux)
      BUILD_LINUX=true
      shift
      ;;
    --windows)
      BUILD_WINDOWS=true
      shift
      ;;
    --clean)
      CLEAN_ONLY=true
      shift
      ;;
    --help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

# Determine what to build
if [ "$CLEAN_ONLY" = true ]; then
  echo "Cleaning distribution binaries..."
  rm -rf bin/duso-* 2>/dev/null || true
  echo "Done"
  exit 0
fi

if [ "$BUILD_ALL" = true ]; then
  BUILD_MACOS=true
  BUILD_LINUX=true
  BUILD_WINDOWS=true
fi

# If no platform selected, build for current platform only (backward compatible)
if [ "$BUILD_MACOS" = false ] && [ "$BUILD_LINUX" = false ] && [ "$BUILD_WINDOWS" = false ]; then
  # Current platform only - use simpler output name
  echo "Building for current platform..."
  go generate ./cmd/duso
  mkdir -p bin
  go build -ldflags "-s -w -X main.Version=$VERSION" -trimpath -o bin/duso ./cmd/duso
  rm -rf cmd/duso/stdlib cmd/duso/docs cmd/duso/contrib cmd/duso/examples
  rm -f cmd/duso/README.md cmd/duso/CONTRIBUTING.md cmd/duso/LICENSE
  echo "✓ Built bin/duso $VERSION"
  exit 0
fi

# Multi-platform build
echo "Generating embedded files..."
go generate ./cmd/duso

mkdir -p bin

BUILT_COUNT=0
FAILED_COUNT=0
ARCHIVE_COUNT=0

# Create temporary directory for staging archives
ARCHIVE_TEMP=$(mktemp -d)
trap "rm -rf $ARCHIVE_TEMP" EXIT

for platform in "${PLATFORMS[@]}"; do
  IFS=':' read -r goos_goarch output_name binary_name <<< "$platform"
  IFS='/' read -r GOOS GOARCH <<< "$goos_goarch"

  # Check if we should build this platform
  skip=false
  case "$GOOS" in
    darwin)
      [ "$BUILD_MACOS" = false ] && skip=true
      ;;
    linux)
      [ "$BUILD_LINUX" = false ] && skip=true
      ;;
    windows)
      [ "$BUILD_WINDOWS" = false ] && skip=true
      ;;
  esac

  if [ "$skip" = true ]; then
    continue
  fi

  echo "Building $GOOS/$GOARCH..."
  if GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -ldflags "-s -w -X main.Version=$VERSION" \
    -trimpath \
    -o "bin/$binary_name" \
    ./cmd/duso; then
    echo "  ✓ bin/$output_name/ (contains $binary_name)"
    ((BUILT_COUNT++))

    # Create archive with binary, LICENSE, and distribution.md
    STAGE_DIR="$ARCHIVE_TEMP/$output_name"
    mkdir -p "$STAGE_DIR"
    cp "bin/$binary_name" "$STAGE_DIR/duso$([ "$GOOS" = "windows" ] && echo ".exe" || echo "")"
    cp LICENSE "$STAGE_DIR/"
    cp docs/distribution.md "$STAGE_DIR/"

    # Create archive
    BIN_DIR="$(cd bin && pwd)"
    if [ "$GOOS" = "windows" ]; then
      # Windows: create zip
      (cd "$ARCHIVE_TEMP" && zip -q -r "$BIN_DIR/${output_name}.zip" "$output_name")
    else
      # Unix: create tar.gz
      (cd "$ARCHIVE_TEMP" && tar czf "$BIN_DIR/${output_name}.tar.gz" "$output_name")
    fi

    if [ $? -eq 0 ]; then
      echo "    ✓ Archive: bin/${output_name}.$([ "$GOOS" = "windows" ] && echo "zip" || echo "tar.gz")"
      ((ARCHIVE_COUNT++))
    fi

    rm -rf "$STAGE_DIR"
  else
    echo "  ✗ FAILED: bin/$output_name/"
    ((FAILED_COUNT++))
  fi
done

# Clean up embedded file copies (they're in .gitignore and will be regenerated on next build)
rm -rf cmd/duso/stdlib cmd/duso/docs cmd/duso/contrib cmd/duso/examples
rm -f cmd/duso/README.md cmd/duso/CONTRIBUTING.md cmd/duso/LICENSE

echo ""
echo "Build summary:"
echo "  Built: $BUILT_COUNT"
echo "  Archives: $ARCHIVE_COUNT"
if [ $FAILED_COUNT -gt 0 ]; then
  echo "  Failed: $FAILED_COUNT"
  exit 1
else
  echo "  ✓ All builds successful"
fi
