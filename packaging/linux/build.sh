#!/bin/bash
# JTNT Agent - Debian Package Builder
# Builds DEB packages for Debian and Ubuntu

set -e

VERSION=${1:-1.0.0}
ARCH=${2:-amd64}
DEB_NAME="jtnt-agent_${VERSION}_${ARCH}.deb"

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

info "JTNT Agent Debian Package Builder"
echo "==================================="
echo ""

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

# Build binaries
info "Building agent for Linux $ARCH..."
cd "$ROOT_DIR"

LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Map architecture
GO_ARCH="$ARCH"
if [ "$ARCH" = "amd64" ]; then
    GO_ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    GO_ARCH="arm64"
elif [ "$ARCH" = "armhf" ]; then
    GO_ARCH="arm"
    export GOARM=7
fi

# Build daemon
info "  Building jtnt-agentd..."
GOOS=linux GOARCH=$GO_ARCH CGO_ENABLED=0 go build \
    -o "$BIN_DIR/jtnt-agentd" \
    -ldflags "$LDFLAGS" \
    -trimpath \
    ./cmd/agentd

success "  jtnt-agentd built"

# Build CLI
info "  Building jtnt-agent..."
GOOS=linux GOARCH=$GO_ARCH CGO_ENABLED=0 go build \
    -o "$BIN_DIR/jtnt-agent" \
    -ldflags "$LDFLAGS" \
    -trimpath \
    ./cmd/jtnt-agent

success "  jtnt-agent built"

# Create package structure
info "Creating package structure..."
DEB_ROOT="$SCRIPT_DIR/deb-root"
rm -rf "$DEB_ROOT"

# Create directory structure
mkdir -p "$DEB_ROOT"/{DEBIAN,usr/local/bin,lib/systemd/system,usr/share/doc/jtnt-agent}

# Copy binaries
cp "$BIN_DIR/jtnt-agentd" "$DEB_ROOT/usr/local/bin/"
cp "$BIN_DIR/jtnt-agent" "$DEB_ROOT/usr/local/bin/"
chmod 755 "$DEB_ROOT/usr/local/bin/jtnt-agentd"
chmod 755 "$DEB_ROOT/usr/local/bin/jtnt-agent"

# Copy systemd unit
cp "$SCRIPT_DIR/debian/jtnt-agentd.service" "$DEB_ROOT/lib/systemd/system/"
chmod 644 "$DEB_ROOT/lib/systemd/system/jtnt-agentd.service"

# Create control file
info "Creating control file..."
sed -e "s/{VERSION}/$VERSION/g" \
    -e "s/{ARCH}/$ARCH/g" \
    "$SCRIPT_DIR/debian/control.template" > "$DEB_ROOT/DEBIAN/control"

# Calculate installed size
INSTALLED_SIZE=$(du -sk "$DEB_ROOT" | cut -f1)
echo "Installed-Size: $INSTALLED_SIZE" >> "$DEB_ROOT/DEBIAN/control"

# Copy maintainer scripts
info "Copying maintainer scripts..."
cp "$SCRIPT_DIR/debian/postinst" "$DEB_ROOT/DEBIAN/"
cp "$SCRIPT_DIR/debian/prerm" "$DEB_ROOT/DEBIAN/"
cp "$SCRIPT_DIR/debian/postrm" "$DEB_ROOT/DEBIAN/"

chmod 755 "$DEB_ROOT/DEBIAN/postinst"
chmod 755 "$DEB_ROOT/DEBIAN/prerm"
chmod 755 "$DEB_ROOT/DEBIAN/postrm"

# Create copyright file
cat > "$DEB_ROOT/usr/share/doc/jtnt-agent/copyright" <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: jtnt-agent
Upstream-Contact: JTNT Communications <support@jtnt.us>
Source: https://github.com/tshojoshua/ai-support-system-agent

Files: *
Copyright: 2025 JTNT Communications
License: Proprietary
 This software is proprietary and licensed for use only with
 the JTNT Hub platform. See End-User License Agreement for details.
EOF

# Create changelog
cat > "$DEB_ROOT/usr/share/doc/jtnt-agent/changelog.Debian.gz" <<EOF
jtnt-agent ($VERSION) stable; urgency=medium

  * Version $VERSION release

 -- JTNT Communications <support@jtnt.us>  $(date -R)
EOF
gzip -9 "$DEB_ROOT/usr/share/doc/jtnt-agent/changelog.Debian.gz"

success "Package structure created"

# Build package
info "Building DEB package..."
if ! command -v dpkg-deb &> /dev/null; then
    error "dpkg-deb not found. Install with: sudo apt-get install dpkg-dev"
    exit 1
fi

dpkg-deb --build --root-owner-group "$DEB_ROOT" "$OUTPUT_DIR/$DEB_NAME"

success "DEB package built"

# Validate package
info "Validating package..."
if command -v lintian &> /dev/null; then
    lintian "$OUTPUT_DIR/$DEB_NAME" || warning "Lintian found some issues (non-fatal)"
else
    warning "lintian not found, skipping validation"
fi

# Clean up
info "Cleaning up..."
rm -rf "$DEB_ROOT"

success "Cleanup complete"

# Get package size
PKG_SIZE=$(du -h "$OUTPUT_DIR/$DEB_NAME" | cut -f1)

# Display package info
if command -v dpkg-deb &> /dev/null; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    success "Build Complete!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Package Details:"
    echo "  File:    $OUTPUT_DIR/$DEB_NAME"
    echo "  Version: $VERSION"
    echo "  Arch:    $ARCH"
    echo "  Size:    $PKG_SIZE"
    echo ""
    echo "Package Info:"
    dpkg-deb --info "$OUTPUT_DIR/$DEB_NAME" | grep -E "(Package|Version|Architecture|Description)" || true
    echo ""
    echo "Installation:"
    echo "  sudo dpkg -i $DEB_NAME"
    echo "  sudo apt-get install -f  # Fix dependencies if needed"
    echo ""
    echo "Or:"
    echo "  sudo apt install ./$DEB_NAME"
    echo ""
    echo "Removal:"
    echo "  sudo apt-get remove jtnt-agent      # Keep data"
    echo "  sudo apt-get purge jtnt-agent       # Remove all data"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi
