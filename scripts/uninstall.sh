#!/bin/bash
# JTNT Agent - Universal Uninstallation Script
# Supports: Linux (Debian/Ubuntu, RHEL/CentOS), macOS

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux*)
        OS_TYPE="linux"
        ;;
    Darwin*)
        OS_TYPE="macos"
        ;;
    *)
        error "Unsupported operating system: $OS"
        ;;
esac

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    error "This script must be run as root. Use: sudo $0"
fi

# Parse arguments
PURGE_DATA=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --purge)
            PURGE_DATA=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            echo "JTNT Agent Uninstaller"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --purge    Remove all agent data (certificates, logs, config)"
            echo "  --force    Skip confirmation prompt"
            echo "  -h, --help Show this help message"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  JTNT Agent Uninstaller"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Confirm uninstallation
if [ "$FORCE" != true ]; then
    echo "This will remove JTNT Agent from your system."
    if [ "$PURGE_DATA" = true ]; then
        echo "WARNING: All agent data (certificates, logs, configuration) will be deleted."
    else
        echo "Agent data will be preserved. Use --purge to remove all data."
    fi
    echo ""
    read -p "Continue? (yes/no): " CONFIRM
    if [ "$CONFIRM" != "yes" ]; then
        echo "Uninstallation cancelled."
        exit 0
    fi
    echo ""
fi

# Uninstall based on OS
case "$OS_TYPE" in
    linux)
        # Detect package manager
        if command -v dpkg &> /dev/null; then
            PKG_MGR="dpkg"
        elif command -v rpm &> /dev/null; then
            PKG_MGR="rpm"
        else
            error "No supported package manager found"
        fi
        
        # Stop service
        info "Stopping JTNT Agent service..."
        if command -v systemctl &> /dev/null; then
            systemctl stop jtnt-agentd 2>/dev/null || true
            systemctl disable jtnt-agentd 2>/dev/null || true
        fi
        success "Service stopped"
        
        # Remove package
        info "Removing package..."
        case "$PKG_MGR" in
            dpkg)
                if [ "$PURGE_DATA" = true ]; then
                    apt-get purge -y jtnt-agent || dpkg --purge jtnt-agent
                else
                    apt-get remove -y jtnt-agent || dpkg --remove jtnt-agent
                fi
                ;;
            rpm)
                if [ "$PURGE_DATA" = true ]; then
                    yum remove -y jtnt-agent || dnf remove -y jtnt-agent || rpm -e jtnt-agent
                    # Manually remove data for RPM
                    rm -rf /var/lib/jtnt-agent
                    rm -rf /etc/jtnt-agent
                    userdel jtnt-agent 2>/dev/null || true
                    groupdel jtnt-agent 2>/dev/null || true
                else
                    yum remove -y jtnt-agent || dnf remove -y jtnt-agent || rpm -e jtnt-agent
                fi
                ;;
        esac
        success "Package removed"
        ;;
        
    macos)
        # Stop and unload service
        info "Stopping JTNT Agent service..."
        PLIST_PATH="/Library/LaunchDaemons/us.jtnt.agentd.plist"
        if [ -f "$PLIST_PATH" ]; then
            launchctl unload -w "$PLIST_PATH" 2>/dev/null || true
            sleep 2
        fi
        
        # Kill any running processes
        pkill -9 jtnt-agentd 2>/dev/null || true
        pkill -9 jtnt-agent 2>/dev/null || true
        success "Service stopped"
        
        # Remove files
        info "Removing installation files..."
        rm -f "$PLIST_PATH"
        rm -rf /usr/local/jtnt/agent
        rmdir /usr/local/jtnt 2>/dev/null || true
        rm -f /etc/paths.d/jtnt-agent
        success "Installation files removed"
        
        # Remove data if purge
        if [ "$PURGE_DATA" = true ]; then
            info "Removing agent data..."
            rm -rf "/Library/Application Support/JTNT/Agent"
            rmdir "/Library/Application Support/JTNT" 2>/dev/null || true
            success "Agent data removed"
        fi
        
        # Remove package receipts
        pkgutil --forget us.jtnt.agent 2>/dev/null || true
        ;;
esac

# Reload systemd if on Linux
if [ "$OS_TYPE" = "linux" ] && command -v systemctl &> /dev/null; then
    systemctl daemon-reload 2>/dev/null || true
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
success "Uninstallation Complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ "$PURGE_DATA" != true ] && [ "$OS_TYPE" = "linux" ]; then
    echo "Agent data preserved in:"
    echo "  /var/lib/jtnt-agent/"
    echo ""
    echo "To remove all data, run:"
    if command -v apt-get &> /dev/null; then
        echo "  sudo apt-get purge jtnt-agent"
    elif command -v yum &> /dev/null; then
        echo "  sudo $0 --purge"
    fi
    echo ""
fi

echo "JTNT Agent has been removed from your system."
echo ""
