# Build script for Duso - PowerShell equivalent to build.sh local binary build

# Ensure we're in the project root
if (-not (Test-Path "go.mod") -or -not (Test-Path "cmd/duso")) {
    Write-Error "Error: build.ps1 must be run from project root"
    exit 1
}

# Get version from git tag or use "dev"
$VERSION = git describe --tags 2>$null
if ($LASTEXITCODE -ne 0) {
    $VERSION = "dev"
}

# Local platform build
Write-Host "Building for current platform..."

# Generate embedded files
go generate ./cmd/duso
if ($LASTEXITCODE -ne 0) {
    Write-Error "go generate failed"
    exit 1
}

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build the binary
$outputPath = "bin/duso.exe"
$ldflags = "-s -w -X main.Version=$VERSION"

go build -ldflags $ldflags -trimpath -o $outputPath ./cmd/duso
if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed"
    exit 1
}

Write-Host "[OK] Built $outputPath $VERSION"
