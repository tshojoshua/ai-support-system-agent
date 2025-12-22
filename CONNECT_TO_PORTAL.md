# Connect to portal.tsho.us for Windows MSI Build

## Quick Start

```bash
# 1. SSH to portal (if SSH is available)
ssh admin@portal.tsho.us

# OR

# 2. RDP to portal
xfreerdp /v:portal.tsho.us /u:admin /size:1920x1080
# (will prompt for password)
```

## If RDP client not installed:

```bash
# Install Remmina (best GUI RDP client for Linux)
sudo apt update
sudo apt install remmina remmina-plugin-rdp -y

# Launch Remmina
remmina
```

In Remmina:
1. Click "+" to add new connection
2. Protocol: RDP
3. Server: `portal.tsho.us`
4. Username: `admin` (or your username)
5. Resolution: 1920x1080 (or custom)
6. Save and connect

## Transfer Agent Code to portal.tsho.us

### Option 1: Git Clone (Easiest)
```powershell
# On portal.tsho.us Windows machine:
cd C:\
git clone https://github.com/tshojoshua/ai-support-system-agent.git jtnt-agent
cd jtnt-agent
```

### Option 2: Transfer Archive
```bash
# On your Linux machine:
cd /home/tsho/ai-support-system/agent

# Create archive
tar czf /tmp/agent-source.tar.gz \
    --exclude='bin' \
    --exclude='.git' \
    --exclude='packaging/*/output' \
    .

# Transfer via SCP (if portal has SSH)
scp /tmp/agent-source.tar.gz admin@portal.tsho.us:C:/temp/

# OR: Start HTTP server and download from Windows
python3 -m http.server 8000
# Then on Windows: http://YOUR_LINUX_IP:8000/agent-source.tar.gz
```

### Option 3: RDP Shared Folder
```bash
# Connect with shared folder
xfreerdp /v:portal.tsho.us /u:admin /drive:agent,/home/tsho/ai-support-system/agent /size:1920x1080

# On Windows, source code will be accessible at:
# \\tsclient\agent
```

## Build MSI on portal.tsho.us

Once connected to portal.tsho.us via RDP:

```powershell
# 1. Open PowerShell as Administrator (right-click, "Run as Administrator")

# 2. Install Chocolatey (if not already installed)
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# 3. Install required tools
choco install -y git golang wix311

# 4. Navigate to agent code
cd C:\jtnt-agent
# (or wherever you cloned/copied the code)

# 5. Build MSI
cd packaging\windows
.\build.ps1 -Version "4.0.0"

# 6. MSI will be in: output\JTNT-Agent-4.0.0-x64.msi
```

## Download MSI from portal.tsho.us

### Option 1: SCP (if SSH available)
```bash
# From your Linux machine
scp admin@portal.tsho.us:C:/jtnt-agent/packaging/windows/output/JTNT-Agent-4.0.0-x64.msi ~/Downloads/
```

### Option 2: RDP Shared Folder
```bash
# Connect with shared folder
xfreerdp /v:portal.tsho.us /u:admin /drive:transfer,/home/tsho/Downloads /size:1920x1080

# On Windows, copy MSI to: \\tsclient\transfer
# It will appear in your Linux ~/Downloads folder
```

### Option 3: Start HTTP Server on Windows
```powershell
# On portal.tsho.us, in PowerShell:
cd C:\jtnt-agent\packaging\windows\output
python -m http.server 8080

# From Linux browser:
# http://portal.tsho.us:8080/JTNT-Agent-4.0.0-x64.msi
```

### Option 4: Copy/Paste (if RDP clipboard enabled)
Small files can be copy/pasted through RDP clipboard sharing.

## Testing on portal.tsho.us

```powershell
# Install test
cd C:\jtnt-agent\packaging\windows\output

# Get test token from Hub
$TEST_TOKEN = "etok_test_xxxxx"

# Silent install
msiexec /i JTNT-Agent-4.0.0-x64.msi /qn /l*v test.log `
  ENROLLMENT_TOKEN=$TEST_TOKEN `
  HUB_URL="https://hub.jtnt.us"

# Wait 30 seconds
Start-Sleep 30

# Check service
Get-Service JTNTAgent

# Check logs
Get-Content "C:\ProgramData\JTNT\Agent\logs\agent.log" -Tail 30

# Uninstall when done
msiexec /x JTNT-Agent-4.0.0-x64.msi /qn
```

## Troubleshooting

### Can't connect to portal.tsho.us
```bash
# Check if host is reachable
ping portal.tsho.us

# Check RDP port open
nc -zv portal.tsho.us 3389

# Try different RDP client
rdesktop portal.tsho.us
```

### Firewall blocking RDP
If you're remote and RDP is blocked:
1. Check if SSH available: `ssh admin@portal.tsho.us`
2. Use VPN if available
3. Ask admin to whitelist your IP

### Build fails on portal
Check [BUILD_GUIDE.md](../packaging/windows/BUILD_GUIDE.md) troubleshooting section.

## Quick Command Summary

```bash
# Connect via RDP
xfreerdp /v:portal.tsho.us /u:admin /drive:agent,/home/tsho/ai-support-system/agent

# On Windows portal:
# PowerShell as Admin
cd C:\jtnt-agent\packaging\windows
.\build.ps1 -Version "4.0.0"

# Output: output\JTNT-Agent-4.0.0-x64.msi
```

That's it! You now have MSI building on portal.tsho.us.
