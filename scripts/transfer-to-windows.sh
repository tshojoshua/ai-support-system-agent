#!/bin/bash
# Transfer agent source code to Windows VM

if [ $# -ne 1 ]; then
    echo "Usage: ./transfer-to-windows.sh <WINDOWS_VM_IP>"
    exit 1
fi

WINDOWS_IP=$1
ADMIN_USER="jtntadmin"

echo "📦 Packaging agent source code..."
cd /home/tsho/ai-support-system/agent
tar czf /tmp/agent-source.tar.gz \
    --exclude='bin' \
    --exclude='*.o' \
    --exclude='.git' \
    --exclude='packaging/*/output' \
    *

echo "📤 Transferring to Windows VM..."
echo "Note: You'll need to install OpenSSH Server on Windows first"
echo "  Or use WinSCP/FileZilla to transfer manually"
echo ""
echo "Archive created: /tmp/agent-source.tar.gz"
echo ""
echo "Manual transfer options:"
echo "1. WinSCP: Download from winscp.net, connect to $WINDOWS_IP"
echo "2. RDP: Connect via RDP and copy/paste or use shared folder"
echo "3. HTTP: python3 -m http.server 8000 (then download from Windows browser)"
echo ""

# Start simple HTTP server for easy download
echo "Starting HTTP server on port 8000..."
echo "On Windows VM, open browser to: http://$(hostname -I | awk '{print $1}'):8000/agent-source.tar.gz"
echo "Press Ctrl+C to stop server"
cd /tmp
python3 -m http.server 8000
