#!/bin/bash
# JTNT Agent - macOS Package Builder
# Builds a universal PKG installer for macOS (Intel + Apple Silicon)

set -e

VERSION=${1:-1.0.0}
PACKAGE_NAME="JTNT-Agent-$VERSION.pkg"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${CYAN}==> ${NC}$1"
}

success() {
    echo -e "${GREEN}✓ ${NC}$1"
}

error() {
    echo -e "${RED}✗ ${NC}$1"
}

warning() {
    echo -e "${YELLOW}⚠ ${NC}$1"
}

info "JTNT Agent macOS Package Builder"
echo "================================="
echo ""

# Validate we're on macOS
if [ "$(uname)" != "Darwin" ]; then
    error "This script must be run on macOS"
    exit 1
fi

# Validate Go is installed
if ! command -v go &> /dev/null; then
    error "Go is not installed. Install from https://go.dev/"
    exit 1
fi

GO_VERSION=$(go version)
success "Go found: $GO_VERSION"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
OUTPUT_DIR="$SCRIPT_DIR/output"

# Create directories
mkdir -p "$BIN_DIR"
mkdir -p "$OUTPUT_DIR"

# Build universal binaries
info "Building universal binaries..."

cd "$ROOT_DIR"

LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Build for Intel (amd64)
info "  Building for Intel (amd64)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
    -o "$BIN_DIR/jtnt-agentd-amd64" \
    -ldflags "$LDFLAGS" \
    -trimpath \
    ./cmd/agentd

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
    -o "$BIN_DIR/jtnt-agent-amd64" \
    -ldflags "$LDFLAGS" \
    -trimpath \
    ./cmd/jtnt-agent

success "  Intel binaries built"

# Build for Apple Silicon (arm64)
info "  Building for Apple Silicon (arm64)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
    -o "$BIN_DIR/jtnt-agentd-arm64" \
    -ldflags "$LDFLAGS" \
    -trimpath \
    ./cmd/agentd

GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
    -o "$BIN_DIR/jtnt-agent-arm64" \
    -ldflags "$LDFLAGS" \
    -trimpath \
    ./cmd/jtnt-agent

success "  Apple Silicon binaries built"

# Create universal binaries with lipo
info "Creating universal binaries..."
lipo -create -output "$BIN_DIR/jtnt-agentd" \
    "$BIN_DIR/jtnt-agentd-amd64" \
    "$BIN_DIR/jtnt-agentd-arm64"

lipo -create -output "$BIN_DIR/jtnt-agent" \
    "$BIN_DIR/jtnt-agent-amd64" \
    "$BIN_DIR/jtnt-agent-arm64"

success "Universal binaries created"

# Verify binaries are universal
info "Verifying universal binaries..."
if lipo -info "$BIN_DIR/jtnt-agentd" | grep -q "x86_64 arm64"; then
    success "  jtnt-agentd is universal"
else
    error "  jtnt-agentd is not universal"
    exit 1
fi

if lipo -info "$BIN_DIR/jtnt-agent" | grep -q "x86_64 arm64"; then
    success "  jtnt-agent is universal"
else
    error "  jtnt-agent is not universal"
    exit 1
fi

# Create package structure
info "Creating package structure..."
PKG_ROOT="$SCRIPT_DIR/pkg-root"
rm -rf "$PKG_ROOT"

mkdir -p "$PKG_ROOT/usr/local/jtnt/agent"
cp "$BIN_DIR/jtnt-agentd" "$PKG_ROOT/usr/local/jtnt/agent/"
cp "$BIN_DIR/jtnt-agent" "$PKG_ROOT/usr/local/jtnt/agent/"
cp "$SCRIPT_DIR/us.jtnt.agentd.plist" "$PKG_ROOT/usr/local/jtnt/agent/"

chmod 755 "$PKG_ROOT/usr/local/jtnt/agent/jtnt-agentd"
chmod 755 "$PKG_ROOT/usr/local/jtnt/agent/jtnt-agent"
chmod 644 "$PKG_ROOT/usr/local/jtnt/agent/us.jtnt.agentd.plist"

success "Package structure created"

# Make scripts executable
chmod +x "$SCRIPT_DIR/scripts/preinstall"
chmod +x "$SCRIPT_DIR/scripts/postinstall"

# Build component package
info "Building component package..."
pkgbuild --root "$PKG_ROOT" \
         --scripts "$SCRIPT_DIR/scripts" \
         --identifier us.jtnt.agent \
         --version "$VERSION" \
         --install-location / \
         "$SCRIPT_DIR/JTNT-Agent-component.pkg"

success "Component package built"

# Create distribution package
info "Building distribution package..."
productbuild --distribution "$SCRIPT_DIR/Distribution.xml" \
             --package-path "$SCRIPT_DIR" \
             --resources "$SCRIPT_DIR/resources" \
             "$OUTPUT_DIR/$PACKAGE_NAME"

success "Distribution package built"

# Sign package if signing identity is set
if [ -n "$SIGNING_IDENTITY" ]; then
    info "Signing package with identity: $SIGNING_IDENTITY"
    
    SIGNED_PKG="$OUTPUT_DIR/JTNT-Agent-$VERSION-signed.pkg"
    
    if productsign --sign "$SIGNING_IDENTITY" \
                   "$OUTPUT_DIR/$PACKAGE_NAME" \
                   "$SIGNED_PKG"; then
        mv "$SIGNED_PKG" "$OUTPUT_DIR/$PACKAGE_NAME"
        success "Package signed successfully"
    else
        warning "Package signing failed, continuing with unsigned package"
    fi
else
    warning "No signing identity set (SIGNING_IDENTITY). Package will be unsigned."
    warning "For notarization, set SIGNING_IDENTITY environment variable."
fi

# Clean up
info "Cleaning up temporary files..."
rm -rf "$PKG_ROOT"
rm -f "$SCRIPT_DIR/JTNT-Agent-component.pkg"
rm -f "$BIN_DIR/jtnt-agentd-amd64" "$BIN_DIR/jtnt-agentd-arm64"
rm -f "$BIN_DIR/jtnt-agent-amd64" "$BIN_DIR/jtnt-agent-arm64"

success "Cleanup complete"

# Get package size
PKG_SIZE=$(du -h "$OUTPUT_DIR/$PACKAGE_NAME" | cut -f1)

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
success "Build Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Package Details:"
echo "  File:     $OUTPUT_DIR/$PACKAGE_NAME"
echo "  Version:  $VERSION"
echo "  Size:     $PKG_SIZE"
echo "  Arch:     Universal (Intel + Apple Silicon)"
echo ""
echo "Installation:"
echo "  Interactive: sudo installer -pkg '$PACKAGE_NAME' -target /"
echo "  CLI:         sudo installer -pkg '$PACKAGE_NAME' -target / -dumplog"
echo ""
echo "Uninstall:"
echo "  sudo $SCRIPT_DIR/uninstall.sh"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
