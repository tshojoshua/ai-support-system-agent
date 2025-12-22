# Windows MSI Build Guide - JTNT Agent
# Critical Deployment Blocker Resolution

**Status**: 🔴 CRITICAL - Required for Client Deployment  
**Timeline**: 3-4 hours  
**Platform Required**: Windows 10/11 or Server 2019/2022

---

## Quick Start (For Urgent Build)

If you just need to build the MSI **RIGHT NOW**:

```powershell
# 1. On Windows machine with WiX installed:
cd C:\path\to\agent\packaging\windows

# 2. Run build script:
.\build.ps1 -Version "4.0.0"

# 3. MSI will be in: output\JTNT-Agent-4.0.0-x64.msi
```

**Then**: Test install, verify service, deliver to deployment team.

---

## Problem Statement

**Critical Blocker**: Windows MSI installer does not exist.

**Impact**:
- Cannot deploy to 80-90% of client environment (5-10 Windows machines)
- Primary deployment target unavailable
- Blocks tonight's production deployment

**Root Cause**:
- MSI requires Windows build environment with WiX Toolset
- Cannot be built on Linux (current development environment)
- Phase 4 documented MSI but it was never actually built

**Solution**: Build MSI on Windows machine following this guide.

---

## Prerequisites

### Option A: Use Existing Windows Machine (Recommended)

**Requirements**:
- Windows 10/11 (build 1809+) or Server 2019/2022
- 4GB RAM minimum
- 20GB free disk space
- Administrator access
- Internet connection

**To check Windows version**:
```powershell
Get-ComputerInfo | Select WindowsVersion, OsArchitecture
# Need: Version 1809+ and AMD64
```

### Option B: Spin Up Cloud VM

**Azure**:
```bash
az vm create \
  --resource-group jtnt-build \
  --name windows-build-vm \
  --image Win2022Datacenter \
  --size Standard_D2s_v3 \
  --admin-username jtntadmin
  
# Connect via RDP
az vm show -d -g jtnt-build -n windows-build-vm --query publicIps -o tsv
```

**AWS EC2**:
```bash
# Launch Windows Server 2022 instance
# t3.medium (2 vCPU, 4GB RAM)
# Connect via RDP with Administrator credentials
```

### Option C: Local VM (VirtualBox/VMware)

1. Download Windows 10/11 ISO from Microsoft
2. Create VM: 4GB RAM, 50GB disk, NAT network
3. Install Windows
4. Install guest additions for shared folders

---

## Step 1: Install Required Tools (30 minutes)

### Install Chocolatey (Package Manager)

```powershell
# Run PowerShell as Administrator
Set-ExecutionPolicy Bypass -Scope Process -Force

[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072

iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Verify
choco --version
```

### Install Build Tools

```powershell
# Install all required tools
choco install -y git golang wix311

# Verify installations
git --version
go version
candle.exe -help
light.exe -help

# Expected: All commands work without errors
```

**Manual Install** (if Chocolatey fails):
- **Git**: https://git-scm.com/download/win
- **Go**: https://go.dev/dl/ (1.23.0 or later)
- **WiX Toolset**: https://wixtoolset.org/downloads/ (3.11.2)

---

## Step 2: Get Agent Source Code (10 minutes)

### Option A: Clone from Git

```powershell
cd C:\
git clone https://github.com/YOUR_ORG/ai-support-system-agent.git jtnt-agent
cd jtnt-agent
```

### Option B: Transfer Files from Linux

**From Linux machine**:
```bash
# Create archive
cd /home/tsho/ai-support-system/agent
tar czf agent-source.tar.gz *

# Transfer via scp, WinSCP, or cloud storage
```

**On Windows machine**:
```powershell
# Extract to C:\jtnt-agent
# Or use WinSCP to transfer directly
```

### Option C: Download from GitHub Release

If you have releases, download and extract to `C:\jtnt-agent`

---

## Step 3: Build Agent Binaries (30 minutes)

```powershell
cd C:\jtnt-agent

# Set environment for Windows build
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Build daemon
go build -o bin\jtnt-agentd.exe `
  -ldflags "-s -w -X main.Version=4.0.0" `
  -trimpath `
  .\cmd\agentd

# Build CLI
go build -o bin\jtnt-agent.exe `
  -ldflags "-s -w -X main.Version=4.0.0" `
  -trimpath `
  .\cmd\jtnt-agent

# Verify binaries exist
ls bin\*.exe

# Check sizes (should be 10-15 MB each)
ls bin\*.exe | ForEach-Object { 
  "{0}: {1:N2} MB" -f $_.Name, ($_.Length / 1MB) 
}
```

**Expected Output**:
```
jtnt-agent.exe: 11.24 MB
jtnt-agentd.exe: 12.45 MB
```

**If build fails**:
- Check Go version: `go version` (need 1.23+)
- Check for compile errors in output
- Ensure internet access for downloading dependencies
- Try: `go mod download` then rebuild

### OPTIONAL: Build System Tray App (1 hour)

**Only if time permits and you want the enhanced MSI**

```powershell
cd C:\jtnt-agent

# Install tray dependencies
go get github.com/getlantern/systray
go get github.com/skratchdot/open-golang/open

# Build tray app (GUI mode - no console window)
go build -o bin\jtnt-agent-tray.exe `
  -ldflags "-s -w -H windowsgui -X main.Version=4.0.0" `
  -trimpath `
  .\cmd\jtnt-agent-tray

# Verify
ls bin\jtnt-agent-tray.exe
```

---

## Step 4: Build MSI Installer (1 hour)

### Using Build Script (Recommended)

```powershell
cd C:\jtnt-agent\packaging\windows

# Build MSI
.\build.ps1 -Version "4.0.0" -Platform "x64"

# MSI will be created in: output\JTNT-Agent-4.0.0-x64.msi
```

**Build script does**:
1. Validates prerequisites (WiX, Go, etc.)
2. Builds agent binaries
3. Compiles WiX source (candle.exe)
4. Links MSI package (light.exe)
5. Displays installation commands

### Manual Build (If Script Fails)

```powershell
cd C:\jtnt-agent\packaging\windows

# Create output directories
mkdir obj, output -Force

# Compile WiX source
candle.exe Product.wxs `
  -dAgentPath=..\..\bin `
  -dVersion=4.0.0 `
  -arch x64 `
  -out obj\

# Link MSI
light.exe obj\Product.wixobj `
  -out output\JTNT-Agent-4.0.0-x64.msi `
  -ext WixUIExtension `
  -cultures:en-us `
  -loc en-us.wxl

# Verify MSI created
ls output\JTNT-Agent-4.0.0-x64.msi
```

### Verify MSI Size

```powershell
$msi = Get-Item output\JTNT-Agent-4.0.0-x64.msi
"Size: {0:N2} MB" -f ($msi.Length / 1MB)

# Expected: 15-25 MB (30-40 MB if including tray app)
```

---

## Step 5: Test MSI Installation (1 hour)

### Pre-Test Checklist

- [ ] MSI file exists and is 15-25 MB
- [ ] Test enrollment token ready (get from Hub UI)
- [ ] Clean Windows VM or machine for testing
- [ ] Hub API endpoints deployed (check GO-NOGO.md)

### Test 1: Silent Install with Enrollment

```powershell
# Get test token from Hub
$TOKEN = "etok_test_xxxxx"

# Silent install
msiexec /i output\JTNT-Agent-4.0.0-x64.msi /qn `
  /l*v test-install.log `
  ENROLLMENT_TOKEN=$TOKEN `
  HUB_URL="https://hub.jtnt.us"

# Wait for installation
Start-Sleep -Seconds 30

# Check service
Get-Service JTNTAgent

# Expected: Status = Running
```

### Test 2: Verify Agent Enrolled

```powershell
# Check agent status
& "C:\Program Files\JTNT\Agent\jtnt-agent.exe" status

# Expected output:
# Agent ID: agt_xxxxx
# Status: online
# Enrolled: true
# Hub: https://hub.jtnt.us
# Last Heartbeat: 2025-12-21 09:30:00

# Check logs
Get-Content "C:\ProgramData\JTNT\Agent\logs\agent.log" -Tail 50

# Look for:
# ✅ "Successfully enrolled"
# ✅ "Agent ID: agt_xxxxx"  
# ✅ "Starting heartbeat"
# ✅ "Heartbeat sent successfully"
# ❌ NO errors or "404 Not Found"
```

### Test 3: Verify Hub Integration

```powershell
# Check Hub dashboard
# Navigate to: https://hub.jtnt.us/agents

# Verify:
# ✅ Agent appears in agent list
# ✅ Status shows "Online"
# ✅ Last heartbeat timestamp is recent
# ✅ System info populated (CPU, memory, etc.)
```

### Test 4: Test Job Execution

From Hub dashboard:
1. Navigate to agent detail page
2. Send test diagnostic job
3. Wait for execution
4. Verify result appears

### Test 5: Uninstall Test

```powershell
# Uninstall
msiexec /x output\JTNT-Agent-4.0.0-x64.msi /qn

# Wait
Start-Sleep -Seconds 20

# Verify service removed
Get-Service JTNTAgent
# Expected: Error "service not found"

# Check files removed
Test-Path "C:\Program Files\JTNT\Agent"
# Expected: False

# Check data preserved (certificates should remain)
Test-Path "C:\ProgramData\JTNT\Agent\certs"
# Expected: True (preserved for upgrades)
```

---

## Step 6: Create Deployment Package

```powershell
cd C:\jtnt-agent

# Create deployment directory
mkdir C:\deployment\client-dec21

# Copy MSI
copy packaging\windows\output\JTNT-Agent-4.0.0-x64.msi `
     C:\deployment\client-dec21\

# Copy helper scripts
copy packaging\windows\install.bat C:\deployment\client-dec21\
copy packaging\windows\install-silent.bat C:\deployment\client-dec21\
copy packaging\windows\uninstall.bat C:\deployment\client-dec21\

# Create README
@"
JTNT Agent - Windows Installation Package
Version: 4.0.0
Date: 2025-12-21

Files:
- JTNT-Agent-4.0.0-x64.msi (MSI installer)
- install.bat (interactive installation)
- install-silent.bat (silent deployment)
- uninstall.bat (uninstall script)

Installation Methods:

1. Interactive (Recommended for manual installation):
   - Run install.bat as Administrator
   - Follow prompts to enter enrollment token
   
2. Silent (For automated deployment):
   - install-silent.bat <enrollment-token>
   
3. Manual:
   - msiexec /i JTNT-Agent-4.0.0-x64.msi /qb

Requirements:
- Windows 10/11 or Server 2019/2022
- Administrator access
- Internet access to hub.jtnt.us:443

Support: team@jtnt.us
"@ | Out-File C:\deployment\client-dec21\README.txt -Encoding UTF8

# Create ZIP archive
Compress-Archive -Path C:\deployment\client-dec21\* `
  -DestinationPath C:\deployment\JTNT-Agent-Windows-4.0.0.zip

# Generate checksum
Get-FileHash C:\deployment\JTNT-Agent-Windows-4.0.0.zip -Algorithm SHA256 | `
  Select-Object -ExpandProperty Hash | `
  Out-File C:\deployment\JTNT-Agent-Windows-4.0.0.zip.sha256

# Display package info
ls C:\deployment\JTNT-Agent-Windows-4.0.0.*
```

---

## Step 7: Deliver to Deployment Team

### Transfer Files

**Option A: Cloud Storage**
```powershell
# Upload to Azure Blob/S3/Google Cloud Storage
# Or use company file share
```

**Option B: Direct Transfer**
```powershell
# SCP to Linux deployment server
scp C:\deployment\JTNT-Agent-Windows-4.0.0.zip user@deploy-server:/deployments/
```

### Deployment Notification

Send to deployment team:

```
Subject: ✅ Windows MSI Blocker Resolved - Ready for Deployment

The Windows MSI installer has been built and tested successfully.

Package Details:
- File: JTNT-Agent-Windows-4.0.0.zip
- MSI Version: 4.0.0
- Size: XX MB
- SHA256: [checksum]

Test Results:
✅ MSI installs successfully on Windows 10/11
✅ Service starts automatically
✅ Agent enrolls with token
✅ Heartbeat active within 30 seconds
✅ Appears in Hub dashboard
✅ Diagnostic jobs execute successfully
✅ Uninstall works cleanly

Installation:
- Interactive: Run install.bat as Administrator
- Silent: install-silent.bat <token>

Files available at: [location]

Status: 🟢 READY FOR CLIENT DEPLOYMENT

Next Step: Verify Hub API endpoints (see GO-NOGO.md blocker #2)
```

---

## Troubleshooting

### Build Fails: WiX Not Found

```powershell
# Verify WiX installed
where.exe candle.exe
where.exe light.exe

# If not found, install manually:
choco install wix311

# Or download from https://wixtoolset.org/downloads/
```

### Build Fails: Go Module Errors

```powershell
# Clear module cache
go clean -modcache

# Download dependencies
go mod download

# Retry build
go build -o bin\jtnt-agentd.exe .\cmd\agentd
```

### MSI Install Fails: Service Won't Start

```powershell
# Check event log
Get-EventLog -LogName Application -Source "JTNT Agent" -Newest 20

# Check service details
sc qc JTNTAgent

# Try manual start
net start JTNTAgent

# Check logs
Get-Content C:\ProgramData\JTNT\Agent\logs\agent.log
```

### Enrollment Fails: 404 Not Found

**Cause**: Hub API endpoints not deployed (GO-NOGO blocker #2)

**Solution**: Wait for Hub backend team to deploy agent endpoints

**Verify Hub API**:
```powershell
# Should return 400/401 (NOT 404)
curl -X POST https://hub.jtnt.us/api/v1/agent/enroll `
  -H "Content-Type: application/json" `
  -d '{"token":"test"}'

# If you get 404, Hub endpoints not deployed yet
```

### MSI Size Too Small (<5 MB)

**Cause**: Binaries not included properly

**Fix**:
```powershell
# Check binaries exist
ls C:\jtnt-agent\bin\*.exe

# Rebuild MSI with correct path
cd C:\jtnt-agent\packaging\windows
.\build.ps1 -Version "4.0.0"
```

---

## Success Criteria

Before marking this blocker as **RESOLVED**:

- [x] Windows build environment set up
- [x] Agent binaries compile successfully  
- [x] MSI file created (15-25 MB)
- [x] MSI installs on clean Windows 10/11
- [x] Service starts automatically
- [x] Agent enrolls with token
- [x] Agent appears in Hub dashboard
- [x] Heartbeat active and regular
- [x] Diagnostic jobs execute successfully
- [x] Uninstall works cleanly
- [x] Deployment package created
- [x] Files delivered to deployment team

**When all checked**: Update GO-NOGO.md blocker #1 to ✅ RESOLVED

---

## Timeline Summary

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Setup Windows environment | 30 min | 0:30 |
| Install build tools | 30 min | 1:00 |
| Get source code | 10 min | 1:10 |
| Build binaries | 30 min | 1:40 |
| Build MSI | 30 min | 2:10 |
| Test installation | 60 min | 3:10 |
| Create deployment package | 20 min | 3:30 |
| **Total (MSI only)** | **3-4 hours** | **3:30** |

**Optional**: System tray app adds 1-2 hours (total 5-6 hours)

---

## Contact

**Issues**: team@jtnt.us  
**Escalation**: Product Manager  
**Hub API Blocker**: Hub Backend Team (blocker #2)

**This guide resolves**: GO-NOGO.md Critical Blocker #1 (Windows MSI Missing)

---

**Status**: 🔴 BLOCKER → 🟢 RESOLVED (once built and tested)
