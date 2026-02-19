# Duso Distribution & Installation Guide

This guide covers distributing Duso binaries and setting up package managers for macOS, Linux, and Windows.

## Quick Release Checklist

- [ ] Build binaries for all architectures:
  - `duso-macos-intel.tar.gz` (macOS x86_64)
  - `duso-macos-silicon.tar.gz` (macOS arm64)
  - `duso-linux-amd64.tar.gz` (Linux x86_64)
  - `duso-linux-arm64.tar.gz` (Linux aarch64)
  - `duso-windows-amd64.zip` (Windows x86_64)
- [ ] Create GitHub release with all binaries attached
- [ ] Update Homebrew formula with new version tag and SHA256 hashes
- [ ] (Optional) Set up Chocolatey for Windows

---

## GitHub Releases

GitHub Releases are your distribution backbone. All package managers fetch binaries from here.

### Upload Binaries

**Via GitHub Web UI:**
1. Go to Releases → Draft new release
2. Tag: `v0.18.19-268` (match your version)
3. Title: `Duso v0.18.19-268`
4. Add release notes
5. Drag/drop or upload all 5 binaries:
   - `duso-macos-intel.tar.gz`
   - `duso-macos-silicon.tar.gz`
   - `duso-linux-amd64.tar.gz`
   - `duso-linux-arm64.tar.gz`
   - `duso-windows-amd64.zip`
6. Publish

**Via CLI:**
```bash
gh release create v0.18.19-268 \
  ./bin/dist/duso-macos-intel.tar.gz \
  ./bin/dist/duso-macos-silicon.tar.gz \
  ./bin/dist/duso-linux-amd64.tar.gz \
  ./bin/dist/duso-linux-arm64.tar.gz \
  ./bin/dist/duso-windows-amd64.zip \
  --title "Duso v0.18.19-268" \
  --notes "Release notes here"
```

### Download URLs

Once released, binaries are available at:
```
https://github.com/duso-org/duso/releases/download/v0.18.19-268/duso-macos-intel.tar.gz
https://github.com/duso-org/duso/releases/download/v0.18.19-268/duso-macos-silicon.tar.gz
https://github.com/duso-org/duso/releases/download/v0.18.19-268/duso-linux-amd64.tar.gz
https://github.com/duso-org/duso/releases/download/v0.18.19-268/duso-linux-arm64.tar.gz
https://github.com/duso-org/duso/releases/download/v0.18.19-268/duso-windows-amd64.zip
```

These URLs are used by package managers and installation scripts.

---

## Homebrew (macOS & Linux)

Homebrew is the fastest and most user-friendly distribution method. One formula covers macOS (Intel & Apple Silicon), Linux (x86_64 & ARM64), and WSL2. The formula automatically selects the correct binary based on the user's architecture.

### Setup (One-Time)

Create a `homebrew-duso` repository in your GitHub org:

```
duso-org/homebrew-duso/
├── Formula/
│   └── duso.rb
└── README.md
```

**Formula:** `Formula/duso.rb` (see [homebrew-duso repo](https://github.com/duso-org/homebrew-duso))
```ruby
class Duso < Formula
  desc "Scripting language for AI agent orchestration"
  homepage "https://github.com/duso-org/duso"
  license "Apache-2.0"

  on_macos do
    on_intel do
      url "https://github.com/duso-org/duso/releases/download/vX.Y.Z/duso-macos-intel.tar.gz"
      sha256 "MACOS_INTEL_SHA256_HERE"
    end
    on_arm do
      url "https://github.com/duso-org/duso/releases/download/vX.Y.Z/duso-macos-silicon.tar.gz"
      sha256 "MACOS_ARM_SHA256_HERE"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/duso-org/duso/releases/download/vX.Y.Z/duso-linux-amd64.tar.gz"
      sha256 "LINUX_AMD64_SHA256_HERE"
    end
    on_arm do
      url "https://github.com/duso-org/duso/releases/download/vX.Y.Z/duso-linux-arm64.tar.gz"
      sha256 "LINUX_ARM64_SHA256_HERE"
    end
  end

  def install
    bin.install "duso"
  end

  test do
    system "#{bin}/duso", "-version"
  end
end
```

### Update for Each Release

1. Upload binaries to GitHub release (see checklist above)

2. Get SHA256 hashes from the GitHub release page

3. Update `homebrew-duso/Formula/duso.rb`:
   - Change version tag: `vX.Y.Z` → `vX.Y.Z+1`
   - Update all 4 `sha256` values (Intel, ARM for macOS + amd64, arm64 for Linux)
   - Commit and push to the [homebrew-duso repo](https://github.com/duso-org/homebrew-duso)

4. Users get automatic updates:
   ```bash
   brew upgrade duso
   ```

### User Installation

```bash
# First time
brew tap duso-org/homebrew-duso
brew install duso

# Later: update
brew upgrade duso

# Check version
duso -version
```

---

## Linux Package Managers (Optional)

### Arch User Repository (AUR)

The Arch community often packages popular tools automatically. You can maintain an official AUR package or let the community handle it.

### Snap

For broad Linux coverage:
```bash
snap install duso
```

Requires `snapcraft.yaml` configuration (15-30 min setup).

### APT / RPM

For Debian/Ubuntu (`.deb`) or RHEL/Fedora (`.rpm`), you'd need to build and maintain packages. Homebrew covers most Linux use cases, so this is optional unless targeting specific distros.

---

## Windows (Optional)

### Chocolatey

Similar to Homebrew, but for Windows:

```powershell
choco install duso
```

Requires a Chocolatey package (similar effort to Homebrew).

### Scoop

Lighter-weight Windows package manager:
```powershell
scoop install duso
```

Even simpler setup than Chocolatey.

### Direct Download

Windows users can always download from GitHub Releases manually.

---

## Installation Script (Fallback)

For users who prefer not to use package managers:

**`install.sh`** (macOS / Linux)
```bash
#!/bin/bash
set -e

BINARY_URL="https://github.com/duso-org/duso/releases/download/v0.16.5/duso-macos"
INSTALL_DIR="/usr/local/bin"

echo "Installing duso..."
curl -fsSL "$BINARY_URL" -o /tmp/duso
chmod +x /tmp/duso
sudo mv /tmp/duso "$INSTALL_DIR/duso"

if ! echo $PATH | grep -q "/usr/local/bin"; then
    echo "export PATH=\"/usr/local/bin:\$PATH\"" >> ~/.zshrc
    echo "Added /usr/local/bin to PATH. Restart your terminal."
fi

echo "✅ duso installed! Try: duso -help"
```

**Usage:**
```bash
curl https://raw.githubusercontent.com/duso-org/duso/main/install.sh | bash
```

---

## Recommended Rollout

**For Release Day (macOS + Linux):**
1. Build and upload binaries to GitHub Releases
2. Set up Homebrew tap
3. Test installation: `brew tap duso-org/homebrew-duso && brew install duso`
4. Document in README: "Install with Homebrew"

**Phase 2 (Optional, post-release):**
- Add Chocolatey package for Windows
- Set up Snap
- Monitor AUR (community usually handles this)

---

## Testing

Before releasing, test the full installation flow:

```bash
# Simulate user installation
brew tap duso-org/homebrew-duso
brew install duso
duso -help
duso -version
```

Verify the binary works as expected and `duso` command is in PATH.

---

## FAQ

**Q: Do I need to support all package managers?**
A: No. Homebrew + GitHub Releases covers ~95% of users. Add others as demand grows.

**Q: How do I update to a new version?**
A: Update the Homebrew formula with new URLs and SHA256 hashes, commit, and users get `brew upgrade duso`.

**Q: Do users need to update their PATH?**
A: No. Package managers handle PATH automatically. Only needed for manual `.tar.gz` downloads.

**Q: Can I distribute a signed binary?**
A: Yes, both macOS (code signing) and Linux (GPG) support it. Recommended for production, but not required for launch.
