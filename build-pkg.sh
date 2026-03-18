#!/bin/bash
set -e

# Build macOS installer package (.pkg)
# Usage: ./build-pkg.sh <intel_binary> <arm64_binary> <output_pkg>

if [ $# -ne 3 ]; then
  echo "Usage: ./build-pkg.sh <intel_binary> <arm64_binary> <output_pkg>"
  echo "Example: ./build-pkg.sh bin/dist/duso-intel bin/dist/duso-arm64 bin/dist/duso.pkg"
  exit 1
fi

INTEL_BIN="$1"
ARM64_BIN="$2"
OUTPUT_PKG="$3"

# Check files exist
if [ ! -f "$INTEL_BIN" ]; then
  echo "Error: Intel binary not found: $INTEL_BIN"
  exit 1
fi

if [ ! -f "$ARM64_BIN" ]; then
  echo "Error: ARM64 binary not found: $ARM64_BIN"
  exit 1
fi

# Create temporary directories
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

PAYLOAD_DIR="$TEMP_DIR/payload/usr/local/bin"
SCRIPTS_DIR="$TEMP_DIR/scripts"
mkdir -p "$PAYLOAD_DIR" "$SCRIPTS_DIR"

# Copy binaries to payload
cp "$INTEL_BIN" "$PAYLOAD_DIR/duso-intel"
cp "$ARM64_BIN" "$PAYLOAD_DIR/duso-silicon"
chmod +x "$PAYLOAD_DIR/duso-intel" "$PAYLOAD_DIR/duso-silicon"

# Create postinstall script
cat > "$SCRIPTS_DIR/postinstall" << 'POSTINSTALL_SCRIPT'
#!/bin/bash
set -e

ARCH=$(uname -m)
INSTALL_DIR="/usr/local/bin"

# Remove existing duso (binary or symlink) if it exists
rm -f "$INSTALL_DIR/duso"

if [ "$ARCH" = "arm64" ]; then
  cp "$INSTALL_DIR/duso-silicon" "$INSTALL_DIR/duso"
elif [ "$ARCH" = "x86_64" ]; then
  cp "$INSTALL_DIR/duso-intel" "$INSTALL_DIR/duso"
else
  echo "Error: Unsupported architecture: $ARCH"
  exit 1
fi

chmod +x "$INSTALL_DIR/duso"
rm -f "$INSTALL_DIR/duso-intel" "$INSTALL_DIR/duso-silicon"

echo "duso installed to $INSTALL_DIR/duso"
POSTINSTALL_SCRIPT

chmod +x "$SCRIPTS_DIR/postinstall"

# Build the package
echo "Creating package: $OUTPUT_PKG"
pkgbuild \
  --root "$TEMP_DIR/payload" \
  --scripts "$SCRIPTS_DIR" \
  --install-location "/" \
  --identifier com.ludonode.duso \
  --version "1.0" \
  "$OUTPUT_PKG"

echo "✓ Package created: $OUTPUT_PKG"
