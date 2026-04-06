#!/bin/bash
set -e

# Ensure we're in the project root
if [ ! -f "go.mod" ] || [ ! -d "cmd/duso" ]; then
  echo "Error: build.sh must be run from project root"
  exit 1
fi

# Load Duso macOS configuration from duso_mac.env if it exists
DUSO_MAC_ENV_PATH="${HOME}/Projects/ludonode/duso_mac.env"
if [ -f "$DUSO_MAC_ENV_PATH" ]; then
  set -a
  source "$DUSO_MAC_ENV_PATH"
  set +a
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
STAPLE_ONLY=false

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
  --staple        Retry stapling the macOS DMG
  --clean         Remove all built binaries
  --help          Show this message

EXAMPLES:
  ./build.sh                # Build for current platform
  ./build.sh --all          # Build for all platforms
  ./build.sh --macos --linux # Build for macOS and Linux only
  ./build.sh --staple       # Retry stapling the macOS DMG
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
    --staple)
      STAPLE_ONLY=true
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

# Function to sign macOS binary with codesign
sign_macos_binary() {
  local binary_path="$1"
  local code_sign_identity="${CODE_SIGN_IDENTITY:-}"

  # If no identity is set, skip signing
  if [ -z "$code_sign_identity" ]; then
    echo "    ⚠ CODE_SIGN_IDENTITY not set, skipping code signing"
    return 0
  fi

  echo "    Signing with codesign..."
  if ! codesign -s "$code_sign_identity" "$binary_path" --force --deep -o runtime; then
    echo "    ✗ Code signing failed"
    return 1
  fi
  echo "    ✓ Code signed"
  return 0
}

# Function to notarize macOS archive with xcrun notarytool
notarize_macos_archive() {
  local archive_path="$1"
  local apple_id="${NOTARIZE_APPLE_ID:-}"
  local team_id="${NOTARIZE_TEAM_ID:-}"
  local password="${NOTARIZE_PASSWORD:-}"

  if [ -z "$apple_id" ]; then
    echo "    ⚠ NOTARIZE_APPLE_ID not set, skipping notarization"
    return 0
  fi

  # Prompt for password if not in env file
  if [ -z "$password" ]; then
    echo "    Enter Apple ID password for notarization:"
    read -s password
    if [ -z "$password" ]; then
      echo "    ⚠ No password provided, skipping notarization"
      return 0
    fi
  fi

  echo "    Notarizing archive with xcrun notarytool..."
  local cmd="xcrun notarytool submit '$archive_path' --apple-id '$apple_id' --password '$password' --wait"
  [ -n "$team_id" ] && cmd="$cmd --team-id '$team_id'"

  if eval "$cmd" > /tmp/notary_request.txt 2>&1; then
    echo "    ✓ Notarization approved, stapling..."

    # Try stapling (may fail if ticket not yet available - use --staple flag to retry later)
    if xcrun stapler staple "$archive_path" 2>/tmp/staple_error.txt; then
      echo "    ✓ Notarization stapled"
      return 0
    else
      echo "    ✗ Stapling failed (run ./build.sh --staple to retry later)"
      return 1
    fi
  else
    echo "    ✗ Notarization submission failed"
    cat /tmp/notary_request.txt
    return 1
  fi
}

# Determine what to build
if [ "$CLEAN_ONLY" = true ]; then
  echo "Cleaning distribution binaries..."
  rm -rf bin/dist 2>/dev/null || true
  echo "Done"
  exit 0
fi

if [ "$STAPLE_ONLY" = true ]; then
  echo "Retrying staple on macOS installer package..."
  PKG_PATH="bin/dist/duso.pkg"
  if [ ! -f "$PKG_PATH" ]; then
    echo "Error: Package not found at $PKG_PATH"
    exit 1
  fi

  # Just retry stapling (reuse the notarize function which includes stapling)
  notarize_macos_archive "$PKG_PATH" && exit 0 || exit 1
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
  rm -rf cmd/duso/app
  mkdir -p cmd/duso/app
  touch cmd/duso/app/PLACEHOLDER
  go generate ./cmd/duso
  mkdir -p bin
  go build -ldflags "-s -w -X main.Version=$VERSION" -trimpath -o bin/duso ./cmd/duso
  # rm -rf cmd/duso/stdlib cmd/duso/docs cmd/duso/contrib cmd/duso/examples
  # rm -f cmd/duso/README.md cmd/duso/CONTRIBUTING.md cmd/duso/LICENSE
  echo "✓ Built bin/duso $VERSION"
  exit 0
fi

# Multi-platform build
echo "Generating embedded files..."
rm -rf cmd/duso/app
mkdir -p cmd/duso/app
touch cmd/duso/app/PLACEHOLDER
go generate ./cmd/duso

mkdir -p bin/dist

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
  # For macOS, build to temp name then rename to avoid overwriting
  BUILD_OUTPUT="bin/dist/$binary_name"
  if [ "$GOOS" = "darwin" ]; then
    BUILD_OUTPUT="bin/dist/duso"
  fi

  if GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -ldflags "-s -w -X main.Version=$VERSION" \
    -trimpath \
    -o "$BUILD_OUTPUT" \
    ./cmd/duso; then
    # For macOS, rename with architecture suffix immediately after building
    if [ "$GOOS" = "darwin" ]; then
      if [ "$GOARCH" = "amd64" ]; then
        mv "$BUILD_OUTPUT" "bin/dist/duso-intel"
        binary_name="duso-intel"
      elif [ "$GOARCH" = "arm64" ]; then
        mv "$BUILD_OUTPUT" "bin/dist/duso-silicon"
        binary_name="duso-silicon"
      fi
    fi
    echo "  ✓ bin/dist/$output_name (contains $binary_name)"
    ((BUILT_COUNT++))

    # Sign macOS binaries
    if [ "$GOOS" = "darwin" ]; then
      sign_macos_binary "bin/dist/$binary_name" || {
        echo "  ✗ Signing failed"
        ((FAILED_COUNT++))
        continue
      }
    fi

    # Create archive with binary, LICENSE, and distribution.md
    STAGE_DIR="$ARCHIVE_TEMP/$output_name"
    mkdir -p "$STAGE_DIR"
    cp "bin/dist/$binary_name" "$STAGE_DIR/duso$([ "$GOOS" = "windows" ] && echo ".exe" || echo "")"
    cp LICENSE "$STAGE_DIR/"
    cp docs/distribution.md "$STAGE_DIR/"

    # Create archive (skip for macOS, will create DMG later)
    if [ "$GOOS" != "darwin" ]; then
      BIN_DIR="$(cd bin/dist && pwd)"
      ARCHIVE_EXT="zip"
      if [ "$GOOS" = "windows" ]; then
        # Windows: create zip
        (cd "$ARCHIVE_TEMP" && zip -q -r "$BIN_DIR/${output_name}.zip" "$output_name")
      else
        # Linux: create tar.gz
        ARCHIVE_EXT="tar.gz"
        (cd "$ARCHIVE_TEMP" && tar czf "$BIN_DIR/${output_name}.tar.gz" "$output_name")
      fi

      if [ $? -eq 0 ]; then
        echo "    ✓ Archive: bin/dist/${output_name}.${ARCHIVE_EXT}"
        ((ARCHIVE_COUNT++))
      fi
    fi

    rm -rf "$STAGE_DIR"
  else
    echo "  ✗ FAILED: bin/dist/$output_name"
    ((FAILED_COUNT++))
  fi
done

# Create macOS universal PKG if we built both architectures
if [ "$BUILD_MACOS" = true ] && [ -f "bin/dist/duso-intel" ] && [ -f "bin/dist/duso-silicon" ]; then
  INTEL_BIN="bin/dist/duso-intel"
  ARM64_BIN="bin/dist/duso-silicon"

  if [ -f "$INTEL_BIN" ] && [ -f "$ARM64_BIN" ]; then
    echo "Creating installer package..."
    BIN_DIR="$(cd bin/dist && pwd)"
    if ./build-pkg.sh "$INTEL_BIN" "$ARM64_BIN" "$BIN_DIR/duso.pkg"; then
      echo "  ✓ Package created: bin/dist/duso.pkg"
      ((ARCHIVE_COUNT++))

      # Sign package with application certificate if set
      PKG_SIGN_IDENTITY="${PKG_SIGN_IDENTITY:-}"
      if [ -n "$PKG_SIGN_IDENTITY" ]; then
        echo "    Signing package with Developer ID Installer..."
        PKG_SIGNED="$BIN_DIR/duso-signed.pkg"
        if productsign --sign "$PKG_SIGN_IDENTITY" "$BIN_DIR/duso.pkg" "$PKG_SIGNED"; then
          mv "$PKG_SIGNED" "$BIN_DIR/duso.pkg"
          echo "    ✓ Package signed"
        else
          echo "    ✗ Package signing failed"
        fi
      fi

      # Notarize package if credentials are set
      notarize_macos_archive "$BIN_DIR/duso.pkg" || {
        echo "  ⚠ Package notarization failed (but package is still usable)"
      }
    else
      echo "  ✗ Package creation failed"
      ((FAILED_COUNT++))
    fi
  fi
fi

# Create individual binary archives (zip files) as fallback
if [ "$BUILD_MACOS" = true ] && [ -f "bin/dist/duso-intel" ] && [ -f "bin/dist/duso-silicon" ]; then
  echo "Creating individual binary archives..."
  ARCHIVE_TEMP=$(mktemp -d)
  trap "rm -rf $ARCHIVE_TEMP" EXIT

  BIN_DIR="$(cd bin/dist && pwd)"

  # Intel archive
  INTEL_STAGE="$ARCHIVE_TEMP/duso-macos-intel"
  mkdir -p "$INTEL_STAGE"
  cp bin/dist/duso-intel "$INTEL_STAGE/duso"
  cp LICENSE "$INTEL_STAGE/"
  cp docs/distribution.md "$INTEL_STAGE/"
  (cd "$ARCHIVE_TEMP" && zip -q -r "$BIN_DIR/duso-macos-intel.zip" "duso-macos-intel")
  echo "  ✓ Archive: bin/dist/duso-macos-intel.zip"

  # Silicon archive
  SILICON_STAGE="$ARCHIVE_TEMP/duso-macos-silicon"
  mkdir -p "$SILICON_STAGE"
  cp bin/dist/duso-silicon "$SILICON_STAGE/duso"
  cp LICENSE "$SILICON_STAGE/"
  cp docs/distribution.md "$SILICON_STAGE/"
  (cd "$ARCHIVE_TEMP" && zip -q -r "$BIN_DIR/duso-macos-silicon.zip" "duso-macos-silicon")
  echo "  ✓ Archive: bin/dist/duso-macos-silicon.zip"

  rm -rf "$ARCHIVE_TEMP"
fi

# Clean up individual binaries from bin/dist, keeping only the archives
for platform in "${PLATFORMS[@]}"; do
  IFS=':' read -r goos_goarch output_name binary_name <<< "$platform"
  IFS='/' read -r GOOS GOARCH <<< "$goos_goarch"

  # Check if we built this platform
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

  if [ "$skip" = false ]; then
    rm -f "bin/dist/$binary_name"
  fi
done

# Clean up macOS architecture-specific binaries
if [ "$BUILD_MACOS" = true ]; then
  rm -f bin/dist/duso-intel bin/dist/duso-silicon
fi

# Clean up embedded file copies (they're in .gitignore and will be regenerated on next build)
# rm -rf cmd/duso/stdlib cmd/duso/docs cmd/duso/contrib cmd/duso/examples
# rm -f cmd/duso/README.md cmd/duso/CONTRIBUTING.md cmd/duso/LICENSE

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
