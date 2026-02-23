# Installing Duso

Choose your preferred installation method below.

## Homebrew (macOS & Linux)

The easiest way to install and update Duso:

```bash
# First time
brew tap duso-org/homebrew-duso
brew install duso

# Later: update
brew upgrade duso
```

## Direct Download from GitHub

Download the binary for your system from [GitHub Releases](https://github.com/duso-org/duso/releases).

### macOS

- **Intel Mac** (Intel processors): `duso-macos-intel.tar.gz`
- **Apple Silicon** (M1, M2, M3, etc.): `duso-macos-silicon.tar.gz`

```bash
# Extract
tar xzf duso-macos-*.tar.gz
cd duso-macos-*

# Install to /usr/local/bin
./duso -install
```

### Linux

- **64-bit x86**: `duso-linux-amd64.tar.gz`
- **ARM64**: `duso-linux-arm64.tar.gz`

```bash
# Extract
tar xzf duso-linux-*.tar.gz
cd duso-linux-*

# Install to /usr/local/bin
./duso -install
```

### Windows

Download `duso-windows-amd64.zip` from [GitHub Releases](https://github.com/duso-org/duso/releases).

```powershell
# Extract
Expand-Archive duso-windows-amd64.zip
cd duso-windows-amd64

# Install to Program Files
.\duso -install
```

The binary will be available as `duso` from any command prompt or PowerShell window.

## Build from Source

You'll need Go 1.21 or later installed. Then:

```bash
git clone https://github.com/duso-org/duso.git
cd duso
./build.sh

# Make it available everywhere
sudo ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

## Verify Installation

```bash
duso -version
```

You should see the version number. If you get "command not found", make sure `/usr/local/bin` is in your PATH:

```bash
echo $PATH
```

If `/usr/local/bin` is missing, add this to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.):

```bash
export PATH="/usr/local/bin:$PATH"
```

Then restart your terminal.

## Uninstall

```bash
# Homebrew
brew uninstall duso
brew untap duso-org/homebrew-duso

# Manual installation
rm /usr/local/bin/duso
```

## License

Apache License 2.0 (see [LICENSE](/LICENSE) file for details) © 2026 Ludonode LLC
