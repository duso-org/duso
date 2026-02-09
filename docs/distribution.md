# Duso Distribution & Installation Guide

This guide covers distributing Duso binaries and setting up package managers for macOS, Linux, and Windows.

## Quick Release Checklist

- [ ] Build binaries: `duso-macos`, `duso-linux`
- [ ] Create GitHub release with binaries attached
- [ ] Set up Homebrew tap (one-time setup)
- [ ] Update Homebrew formula with new version URLs
- [ ] (Optional) Set up Chocolatey for Windows

---

## GitHub Releases

GitHub Releases are your distribution backbone. All package managers fetch binaries from here.

### Upload Binaries

**Via GitHub Web UI:**
1. Go to Releases → Draft new release
2. Tag: `v0.16.5` (match your version)
3. Title: `Duso v0.16.5`
4. Add release notes
5. Drag/drop or upload:
   - `duso-macos.tar.gz` (or `.zip`)
   - `duso-linux.tar.gz` (or `.zip`)
6. Publish

**Via CLI:**
```bash
gh release create v0.16.5 \
  ./bin/duso-macos.tar.gz \
  ./bin/duso-linux.tar.gz \
  --title "Duso v0.16.5" \
  --notes "Release notes here"
```

### Download URLs

Once released, binaries are available at:
```
https://github.com/duso-org/duso/releases/download/v0.16.5/duso-macos.tar.gz
https://github.com/duso-org/duso/releases/download/v0.16.5/duso-linux.tar.gz
```

These URLs are used by package managers and installation scripts.

---

## Homebrew (macOS & Linux)

Homebrew is the fastest and most user-friendly distribution method. One formula covers macOS, Linux (via Linuxbrew), and WSL2.

### Setup (One-Time)

Create a `homebrew-duso` repository in your GitHub org:

```
duso-org/homebrew-duso/
├── Formula/
│   └── duso.rb
└── README.md
```

**Formula:** `Formula/duso.rb`
```ruby
class Duso < Formula
  desc "Scripting language for AI agent orchestration"
  homepage "https://github.com/duso-org/duso"
  license "Apache-2.0"

  on_macos do
    url "https://github.com/duso-org/duso/releases/download/v0.16.5/duso-macos.tar.gz"
    sha256 "MACOS_SHA256_HERE"
  end

  on_linux do
    url "https://github.com/duso-org/duso/releases/download/v0.16.5/duso-linux.tar.gz"
    sha256 "LINUX_SHA256_HERE"
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

1. Get SHA256 hashes:
   ```bash
   sha256sum duso-macos.tar.gz
   sha256sum duso-linux.tar.gz
   ```

2. Update `Formula/duso.rb`:
   - Change version tag: `v0.16.5` → `v0.16.6`
   - Update both `sha256` values
   - Commit and push

3. Users get automatic updates:
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
