# JTNT Agent - Build MSI Installer
# Requires: WiX Toolset 3.11+ or WiX 4.0+
# Usage: .\build.ps1 [-Version "1.0.0"] [-Platform "x64"]

param(
    [string]$Version = "1.0.0",
    [string]$Platform = "x64",
    [string]$Configuration = "Release"
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Colors for output
function Write-Status {
    param([string]$Message)
    Write-Host "==> " -ForegroundColor Cyan -NoNewline
    Write-Host $Message
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-ErrorMsg {
    param([string]$Message)
    Write-Host "✗ " -ForegroundColor Red -NoNewline
    Write-Host $Message
}

# Validate WiX is installed
Write-Status "Checking for WiX Toolset..."
$wixPath = Get-Command candle.exe -ErrorAction SilentlyContinue
if (-not $wixPath) {
    Write-ErrorMsg "WiX Toolset not found. Please install from https://wixtoolset.org/"
    exit 1
}
Write-Success "WiX Toolset found: $($wixPath.Source)"

# Validate Go is installed
Write-Status "Checking for Go..."
$goPath = Get-Command go.exe -ErrorAction SilentlyContinue
if (-not $goPath) {
    Write-ErrorMsg "Go not found. Please install from https://go.dev/"
    exit 1
}
$goVersion = go version
Write-Success "Go found: $goVersion"

# Create output directories
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent (Split-Path -Parent $scriptDir)
$binDir = Join-Path $rootDir "bin"
$objDir = Join-Path $scriptDir "obj"
$outputDir = Join-Path $scriptDir "output"

New-Item -ItemType Directory -Force -Path $binDir | Out-Null
New-Item -ItemType Directory -Force -Path $objDir | Out-Null
New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Write-Status "Output directories created"

# Build agent binaries
Write-Status "Building agent daemon for Windows $Platform..."
Push-Location $rootDir
try {
    $env:GOOS = "windows"
    $env:GOARCH = if ($Platform -eq "x64") { "amd64" } else { $Platform }
    $env:CGO_ENABLED = "0"
    
    $ldflags = "-s -w -X main.Version=$Version -X main.BuildTime=$(Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ')"
    
    # Build daemon
    Write-Status "  Building jtnt-agentd.exe..."
    go build -o "$binDir\jtnt-agentd.exe" -ldflags $ldflags -trimpath ./cmd/agentd
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build jtnt-agentd.exe"
    }
    Write-Success "  jtnt-agentd.exe built successfully"
    
    # Build CLI
    Write-Status "  Building jtnt-agent.exe..."
    go build -o "$binDir\jtnt-agent.exe" -ldflags $ldflags -trimpath ./cmd/jtnt-agent
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build jtnt-agent.exe"
    }
    Write-Success "  jtnt-agent.exe built successfully"
    
} finally {
    Pop-Location
    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
}

# Get file version from binary
$agentExe = Join-Path $binDir "jtnt-agentd.exe"
$fileVersion = (Get-Item $agentExe).VersionInfo.FileVersion
if (-not $fileVersion) {
    $fileVersion = $Version
}
Write-Status "Binary version: $fileVersion"

# Build WiX project
Write-Status "Building MSI installer..."
$msiName = "JTNT-Agent-$Version-$Platform.msi"
$msiPath = Join-Path $outputDir $msiName

try {
    # Compile WiX source
    Write-Status "  Running candle.exe (compile)..."
    $candleArgs = @(
        "-dAgentPath=$binDir"
        "-dVersion=$Version"
        "-arch", $Platform
        "-out", "$objDir\"
        "Product.wxs"
    )
    
    Push-Location $scriptDir
    & candle.exe $candleArgs
    if ($LASTEXITCODE -ne 0) {
        throw "candle.exe failed with exit code $LASTEXITCODE"
    }
    Write-Success "  WiX compilation successful"
    
    # Link WiX objects
    Write-Status "  Running light.exe (link)..."
    $lightArgs = @(
        "-out", $msiPath
        "-ext", "WixUIExtension"
        "-cultures:en-us"
        "-loc", "en-us.wxl"
        "$objDir\Product.wixobj"
    )
    
    & light.exe $lightArgs 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        # Light.exe returns warnings as exit code, check if MSI exists
        if (-not (Test-Path $msiPath)) {
            throw "light.exe failed - MSI not created"
        }
        Write-Host "  (warnings ignored)" -ForegroundColor Yellow
    }
    Write-Success "  MSI linking successful"
    
} finally {
    Pop-Location
}

# Validate MSI was created
if (-not (Test-Path $msiPath)) {
    Write-ErrorMsg "MSI file not found: $msiPath"
    exit 1
}

# Get MSI file size
$msiSize = (Get-Item $msiPath).Length / 1MB
Write-Success "MSI created successfully"
Write-Host ""
Write-Host "Installer Details:" -ForegroundColor Cyan
Write-Host "  File:     $msiPath"
Write-Host "  Version:  $Version"
Write-Host "  Platform: $Platform"
Write-Host "  Size:     $([math]::Round($msiSize, 2)) MB"
Write-Host ""
Write-Host "Installation Commands:" -ForegroundColor Cyan
Write-Host "  Interactive: " -NoNewline
Write-Host "msiexec /i `"$msiName`"" -ForegroundColor Yellow
Write-Host "  Silent:      " -NoNewline
Write-Host "msiexec /i `"$msiName`" /qn" -ForegroundColor Yellow
Write-Host "  With Token:  " -NoNewline
Write-Host "msiexec /i `"$msiName`" /qn ENROLLMENT_TOKEN=`"your-token`" HUB_URL=`"https://hub.jtnt.us`"" -ForegroundColor Yellow
Write-Host ""

# Cleanup object files
Write-Status "Cleaning up temporary files..."
Remove-Item -Recurse -Force $objDir -ErrorAction SilentlyContinue
Write-Success "Build complete!"
