# Installing Duso

## macOS Package Installer

The fastest way to get Duso on macOS. Download the signed installer from [duso.rocks/download](https://duso.rocks/download) and double-click `duso-installer.pkg`. Works on Intel and Apple Silicon.

## Manual Install

Available for:

- **macOS** - Intel and Apple Silicon
- **Linux x86_64** - AMD64 architecture
- **Linux ARM64** - ARM-based Linux
- **Windows 10+** - x86_64 architecture

### Download from the Web

Visit [duso.rocks/download](https://duso.rocks/download), download the version for your system, and follow the instructions on the page.

### Command Line Download

If you prefer to download from the terminal, use the direct GitHub release URLs:

**macOS (Intel):**

```bash
curl -LO https://github.com/duso-org/duso/releases/download/v1.0.8-400/duso-macos-intel.tar.gz
tar xzf duso-macos-intel.tar.gz && cd duso-macos-intel && ./duso install
```

**macOS (Apple Silicon):**

```bash
curl -LO https://github.com/duso-org/duso/releases/download/v1.0.8-400/duso-macos-silicon.tar.gz
tar xzf duso-macos-silicon.tar.gz && cd duso-macos-silicon && ./duso install
```

**Linux x86_64:**

```bash
curl -LO https://github.com/duso-org/duso/releases/download/v1.0.8-400/duso-linux-amd64.tar.gz
tar xzf duso-linux-amd64.tar.gz && cd duso-linux-amd64 && ./duso install
```

**Linux ARM64:**

```bash
curl -LO https://github.com/duso-org/duso/releases/download/v1.0.8-400/duso-linux-arm64.tar.gz
tar xzf duso-linux-arm64.tar.gz && cd duso-linux-arm64 && ./duso install
```

**Windows:**

```bash
curl -LO https://github.com/duso-org/duso/releases/download/v1.0.8-400/duso-windows-amd64.zip
Expand-Archive duso-windows-amd64.zip && cd duso-windows-amd64 && .\duso install
```

## Homebrew (macOS & Linux)

For easy updates:

```bash
brew tap duso-org/homebrew-duso
brew install duso
```

Later, update with:

```bash
brew upgrade duso
```

### Uninstall

```bash
brew uninstall duso
brew untap duso-org/homebrew-duso
```

## Build from Source

Requires Go 1.25 or later:

```bash
git clone https://github.com/duso-org/duso.git
cd duso
./build.sh

# Make it available everywhere
sudo ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

## Verify Installation

```bash
duso
```

If you get "command not found", make sure `/usr/local/bin` is in your PATH. Add this to your shell profile if needed:

```bash
export PATH="/usr/local/bin:$PATH"
```
