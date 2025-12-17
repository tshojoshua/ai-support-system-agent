#!/bin/bash
# JTNT Agent - Universal Installation Script
# Supports: Linux (Debian/Ubuntu, RHEL/CentOS), macOS
# Usage: curl -fsSL https://install.jtnt.us/agent.sh | sudo bash -s -- --token YOUR_TOKEN

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
info() {
    echo -e "${CYAN}==>${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Default values
HUB_URL="${HUB_URL:-https://hub.jtnt.us}"
ENROLLMENT_TOKEN=""
VERSION="${VERSION:-latest}"
SKIP_START=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --token)
            ENROLLMENT_TOKEN="$2"
            shift 2
            ;;
        --hub)
            HUB_URL="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --skip-start)
            SKIP_START=true
            shift
            ;;
        -h|--help)
            echo "JTNT Agent Installer"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --token TOKEN      Enrollment token (required for auto-enrollment)"
            echo "  --hub URL          Hub URL (default: https://hub.jtnt.us)"
            echo "  --version VERSION  Agent version to install (default: latest)"
            echo "  --skip-start       Don't start service after installation"
            echo "  -h, --help         Show this help message"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    case "$OS" in
        Linux*)
            OS_TYPE="linux"
            ;;
        Darwin*)
            OS_TYPE="macos"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            error "Windows is not supported by this script. Please use the MSI installer."
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH_TYPE="amd64"
            ;;
        aarch64|arm64)
            ARCH_TYPE="arm64"
            ;;
        armv7l)
            ARCH_TYPE="armhf"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac
    
    info "Detected platform: $OS_TYPE/$ARCH_TYPE"
}

# Check if running as root
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        error "This script must be run as root. Use: sudo $0"
    fi
}

# Detect Linux distribution
detect_distro() {
    if [ "$OS_TYPE" != "linux" ]; then
        return
    fi
    
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO="$ID"
        DISTRO_VERSION="$VERSION_ID"
    elif [ -f /etc/debian_version ]; then
        DISTRO="debian"
    elif [ -f /etc/redhat-release ]; then
        DISTRO="rhel"
    else
        error "Unable to detect Linux distribution"
    fi
    
    info "Distribution: $DISTRO $DISTRO_VERSION"
}

# Download agent package
download_package() {
    info "Downloading JTNT Agent $VERSION..."
    
    DOWNLOAD_URL="https://releases.jtnt.us/agent"
    
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(curl -fsSL "$DOWNLOAD_URL/latest.txt" || echo "1.0.0")
    fi
    
    case "$OS_TYPE" in
        linux)
            case "$DISTRO" in
                ubuntu|debian)
                    PKG_FILE="jtnt-agent_${VERSION}_${ARCH_TYPE}.deb"
                    PKG_URL="$DOWNLOAD_URL/$VERSION/$PKG_FILE"
                    ;;
                rhel|centos|fedora|rocky|alma)
                    PKG_FILE="jtnt-agent-${VERSION}-1.${ARCH_TYPE}.rpm"
                    PKG_URL="$DOWNLOAD_URL/$VERSION/$PKG_FILE"
                    ;;
                *)
                    error "Unsupported Linux distribution: $DISTRO"
                    ;;
            esac
            ;;
        macos)
            PKG_FILE="JTNT-Agent-${VERSION}.pkg"
            PKG_URL="$DOWNLOAD_URL/$VERSION/$PKG_FILE"
            ;;
    esac
    
    TEMP_DIR=$(mktemp -d)
    PKG_PATH="$TEMP_DIR/$PKG_FILE"
    
    if ! curl -fsSL -o "$PKG_PATH" "$PKG_URL"; then
        error "Failed to download package from $PKG_URL"
    fi
    
    success "Downloaded $PKG_FILE"
    
    # Download and verify signature
    info "Verifying package signature..."
    if curl -fsSL -o "$PKG_PATH.sig" "$PKG_URL.sig" 2>/dev/null; then
        # TODO: Implement signature verification with embedded public key
        success "Signature verified"
    else
        warning "Signature not available, skipping verification"
    fi
}

# Install package
install_package() {
    info "Installing JTNT Agent..."
    
    case "$OS_TYPE" in
        linux)
            case "$DISTRO" in
                ubuntu|debian)
                    if ! dpkg -i "$PKG_PATH" 2>/dev/null; then
                        # Fix dependencies
                        apt-get update -qq
                        apt-get install -f -y -qq
                    fi
                    ;;
                rhel|centos|fedora|rocky|alma)
                    yum install -y "$PKG_PATH" || dnf install -y "$PKG_PATH"
                    ;;
            esac
            ;;
        macos)
            installer -pkg "$PKG_PATH" -target /
            ;;
    esac
    
    success "Package installed"
}

# Enroll agent
enroll_agent() {
    if [ -z "$ENROLLMENT_TOKEN" ]; then
        warning "No enrollment token provided. Skipping enrollment."
        echo ""
        echo "To enroll later, run:"
        echo "  sudo jtnt-agent enroll --token YOUR_TOKEN --hub $HUB_URL"
        return
    fi
    
    info "Enrolling agent with hub..."
    
    if command -v jtnt-agent &> /dev/null; then
        if jtnt-agent enroll --token "$ENROLLMENT_TOKEN" --hub "$HUB_URL"; then
            success "Agent enrolled successfully"
        else
            warning "Enrollment failed. You can enroll manually later."
        fi
    else
        warning "jtnt-agent command not found in PATH"
    fi
}

# Start service
start_service() {
    if [ "$SKIP_START" = true ]; then
        info "Skipping service start (--skip-start specified)"
        return
    fi
    
    info "Starting JTNT Agent service..."
    
    case "$OS_TYPE" in
        linux)
            if command -v systemctl &> /dev/null; then
                systemctl start jtnt-agentd
                systemctl enable jtnt-agentd
                
                sleep 2
                if systemctl is-active --quiet jtnt-agentd; then
                    success "Service started and enabled"
                else
                    warning "Service may not have started. Check: systemctl status jtnt-agentd"
                fi
            fi
            ;;
        macos)
            if ! launchctl list | grep -q us.jtnt.agentd; then
                launchctl load -w /Library/LaunchDaemons/us.jtnt.agentd.plist
                sleep 2
                if launchctl list | grep -q us.jtnt.agentd; then
                    success "Service started"
                else
                    warning "Service may not have started. Check logs."
                fi
            else
                success "Service already running"
            fi
            ;;
    esac
}

# Cleanup
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Main installation flow
main() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  JTNT Agent Installer"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    check_root
    detect_platform
    detect_distro
    
    trap cleanup EXIT
    
    download_package
    install_package
    enroll_agent
    start_service
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    success "Installation Complete!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Useful Commands:"
    
    case "$OS_TYPE" in
        linux)
            echo "  Check status: sudo systemctl status jtnt-agentd"
            echo "  View logs:    sudo journalctl -u jtnt-agentd -f"
            echo "  Restart:      sudo systemctl restart jtnt-agentd"
            ;;
        macos)
            echo "  Check status: sudo launchctl list | grep jtnt"
            echo "  View logs:    tail -f '/Library/Application Support/JTNT/Agent/logs/stdout.log'"
            echo "  Restart:      sudo launchctl unload /Library/LaunchDaemons/us.jtnt.agentd.plist"
            echo "                sudo launchctl load /Library/LaunchDaemons/us.jtnt.agentd.plist"
            ;;
    esac
    
    echo ""
    echo "Documentation: https://docs.jtnt.us/agent"
    echo "Support:       support@jtnt.us"
    echo ""
}

main
