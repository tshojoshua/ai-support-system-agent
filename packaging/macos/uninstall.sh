#!/bin/bash
# JTNT Agent - macOS Uninstaller
# Complete removal of JTNT Agent from the system

set -e

# Check for root privileges
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root. Use: sudo $0"
    exit 1
fi

echo "JTNT Agent Uninstaller"
echo "======================"
echo ""

# Confirm uninstallation
read -p "This will completely remove JTNT Agent. Continue? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Uninstallation cancelled."
    exit 0
fi

echo ""

# Stop and unload service
PLIST_PATH="/Library/LaunchDaemons/us.jtnt.agentd.plist"
if [ -f "$PLIST_PATH" ]; then
    echo "Stopping JTNT Agent service..."
    launchctl unload -w "$PLIST_PATH" 2>/dev/null || true
    sleep 2
    echo "✓ Service stopped"
fi

# Kill any running processes
echo "Terminating any running agent processes..."
pkill -9 jtnt-agentd 2>/dev/null || true
pkill -9 jtnt-agent 2>/dev/null || true
echo "✓ Processes terminated"

# Remove launchd plist
if [ -f "$PLIST_PATH" ]; then
    echo "Removing service configuration..."
    rm -f "$PLIST_PATH"
    echo "✓ Service configuration removed"
fi

# Remove installation directory
INSTALL_DIR="/usr/local/jtnt/agent"
if [ -d "$INSTALL_DIR" ]; then
    echo "Removing installation files..."
    rm -rf "$INSTALL_DIR"
    echo "✓ Installation files removed"
fi

# Remove parent directory if empty
if [ -d "/usr/local/jtnt" ]; then
    rmdir "/usr/local/jtnt" 2>/dev/null || true
fi

# Ask about state directory
STATE_DIR="/Library/Application Support/JTNT/Agent"
if [ -d "$STATE_DIR" ]; then
    echo ""
    read -p "Remove agent data (certificates, logs, config)? (yes/no): " REMOVE_DATA
    if [ "$REMOVE_DATA" = "yes" ]; then
        echo "Removing agent data..."
        rm -rf "$STATE_DIR"
        echo "✓ Agent data removed"
        
        # Remove parent directory if empty
        if [ -d "/Library/Application Support/JTNT" ]; then
            rmdir "/Library/Application Support/JTNT" 2>/dev/null || true
        fi
    else
        echo "Agent data preserved at: $STATE_DIR"
    fi
fi

# Remove from PATH
PATHS_FILE="/etc/paths.d/jtnt-agent"
if [ -f "$PATHS_FILE" ]; then
    echo "Removing from system PATH..."
    rm -f "$PATHS_FILE"
    echo "✓ PATH updated"
fi

# Remove receipts (package database)
echo "Cleaning up package receipts..."
pkgutil --forget us.jtnt.agent 2>/dev/null || true
echo "✓ Package receipts removed"

echo ""
echo "Uninstallation Complete"
echo "======================="
echo "JTNT Agent has been removed from this system."
echo ""

exit 0
