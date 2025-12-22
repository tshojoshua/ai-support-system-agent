# JTNT Agent - Deployment & Testing Guide

## Built Installers

### ✅ Linux DEB Package (Ready to Test)
**Location:** `packaging/linux/output/jtnt-agent_1.0.0_amd64.deb`
**Size:** 608KB
**Architecture:** amd64 (x86_64)

### ⚠️ Windows MSI & macOS PKG (Not Yet Built)
- Requires platform-specific build tools
- Can be built via GitHub Actions workflow
- Or built manually on respective platforms

---

## Linux Testing (Current System)

### 1. Install the DEB Package

```bash
# Install using dpkg
sudo dpkg -i packaging/linux/output/jtnt-agent_1.0.0_amd64.deb

# Or use apt (recommended - handles dependencies)
sudo apt install ./packaging/linux/output/jtnt-agent_1.0.0_amd64.deb
```

### 2. Verify Installation

```bash
# Check if binaries are installed
which jtnt-agent
which jtnt-agentd

# Test daemon
jtnt-agentd --version

# Test CLI
jtnt-agent --version

# Check systemd service
systemctl status jtnt-agentd
```

### 3. Check Created Files/Directories

```bash
# Service account
id jtnt-agent

# State directories
ls -la /var/lib/jtnt-agent/
ls -la /etc/jtnt-agent/

# systemd unit
systemctl cat jtnt-agentd

# Binary locations
ls -l /usr/local/bin/jtnt-agent*
```

### 4. Test Enrollment (Minimal Build)

```bash
# Test enrollment command (won't actually connect)
sudo jtnt-agent enroll --token test-token-123 --hub https://hub.jtnt.us
```

### 5. Test Service

```bash
# Start service
sudo systemctl start jtnt-agentd

# Check status
sudo systemctl status jtnt-agentd

# View logs
sudo journalctl -u jtnt-agentd -f

# Stop service
sudo systemctl stop jtnt-agentd
```

### 6. Uninstall Testing

```bash
# Remove package (keeps data)
sudo apt-get remove jtnt-agent

# Verify service stopped
systemctl status jtnt-agentd

# Check if data preserved
ls -la /var/lib/jtnt-agent/
ls -la /etc/jtnt-agent/

# Reinstall
sudo apt install ./packaging/linux/output/jtnt-agent_1.0.0_amd64.deb

# Complete purge (removes all data)
sudo apt-get purge jtnt-agent

# Verify complete removal
ls /var/lib/jtnt-agent/  # Should not exist
id jtnt-agent  # Should not exist
```

---

## Windows Testing (Requires Windows Machine)

### Prerequisites
- Windows 10+
- Administrator privileges

### Build MSI (On Windows)
```powershell
.\packaging\windows\build.ps1 -Version 1.0.0
```

### Install
1. Double-click `jtnt-agent-1.0.0-x64.msi`
2. Follow installation wizard
3. Or silent install:
   ```powershell
   msiexec /i jtnt-agent-1.0.0-x64.msi /quiet /qn
   ```

### Verify
```powershell
# Check service
sc query jtnt-agentd

# Test binaries
& "C:\Program Files\JTNT\Agent\jtnt-agent.exe" --version
& "C:\Program Files\JTNT\Agent\jtnt-agentd.exe" --version

# Check installation directory
dir "C:\Program Files\JTNT\Agent\"
dir "C:\ProgramData\JTNT\Agent\"
```

### Uninstall
```powershell
# Via Control Panel or
msiexec /x jtnt-agent-1.0.0-x64.msi /quiet
```

---

## macOS Testing (Requires macOS Machine)

### Prerequisites
- macOS 13.0+ (Ventura)
- Xcode Command Line Tools
- Admin privileges

### Build PKG (On macOS)
```bash
./packaging/macos/build.sh 1.0.0
```

### Install
```bash
# GUI installer
sudo installer -pkg packaging/macos/output/jtnt-agent-1.0.0.pkg -target /

# Or double-click PKG file
```

### Verify
```bash
# Check LaunchDaemon
sudo launchctl list | grep jtnt

# Test binaries
/usr/local/jtnt/agent/jtnt-agent --version
/usr/local/jtnt/agent/jtnt-agentd --version

# Check installation
ls -la /usr/local/jtnt/agent/
ls -la "/Library/Application Support/JTNT/Agent/"
```

### Uninstall
```bash
sudo /usr/local/jtnt/agent/uninstall.sh
```

---

## Cross-Platform Testing via GitHub Actions

### Automatic Builds
The `.github/workflows/build.yml` workflow builds all platforms:

1. Push changes to GitHub:
   ```bash
   git add .
   git commit -m "Build test installers"
   git push origin main
   ```

2. GitHub Actions will automatically:
   - Build Windows MSI
   - Build macOS PKG (Intel + Apple Silicon)
   - Build Linux DEB (amd64, arm64, armhf)
   - Sign packages
   - Create GitHub Release
   - Upload artifacts

3. Download installers from:
   - GitHub Actions artifacts tab
   - GitHub Releases page

---

## Current Build Status

### ✅ What Works
- Linux DEB package builds successfully
- Minimal binaries (daemon + CLI) work
- Package installation/removal tested
- systemd integration tested
- User/group creation tested
- Directory permissions tested

### ⚠️ Known Limitations (Minimal Build)
- Full agent functionality disabled due to compilation errors
- Enrollment connects but doesn't complete
- Job execution not available
- Metrics/monitoring not available

### 🔧 To Enable Full Functionality
The main Go code has compilation errors that need fixing:

1. **Agent struct missing fields:**
   - `mu sync.Mutex`
   - `currentJob *api.Job`
   - `jobPollingStopped chan struct{}`
   - `metricsServer *http.Server`

2. **Files to fix:**
   - `internal/agent/agent.go` - Add missing struct fields
   - `internal/agent/shutdown.go` - Uses undefined fields
   - `internal/agent/heartbeat.go` - ticker.Reset() usage
   - `internal/agent/job_loop.go` - Job field name mismatches

3. **Quick fix** (restore broken files and fix):
   ```bash
   mv cmd/agentd/main_broken.go cmd/agentd/main.go
   mv cmd/jtnt-agent/main_broken.go cmd/jtnt-agent/main.go
   # Then fix compilation errors in internal/agent
   ```

---

## Test Scenarios

### Scenario 1: Fresh Install
1. Install package
2. Verify service created but not running
3. Run enrollment command
4. Start service
5. Check logs

### Scenario 2: Upgrade
1. Install version 1.0.0
2. Create some test data in `/var/lib/jtnt-agent/`
3. Install version 1.0.1 (when available)
4. Verify data preserved
5. Verify service restarted

### Scenario 3: Uninstall
1. Install package
2. Create test data
3. Remove package (apt remove)
4. Verify service stopped
5. Verify data preserved
6. Purge package (apt purge)
7. Verify complete removal

### Scenario 4: Service Management
1. Start service
2. Check it's running
3. Kill process manually
4. Verify systemd restarts it
5. Stop via systemctl
6. Verify it stays stopped

---

## Production Deployment

### For Production Use:
1. Fix compilation errors (restore full agent code)
2. Build with proper version numbers
3. Sign packages with Ed25519 keys
4. Test enrollment with real hub
5. Deploy via:
   - Direct package installation
   - Configuration management (Ansible, Chef, Puppet)
   - Universal install script:
     ```bash
     curl -fsSL https://get.jtnt.us/install.sh | sudo bash -s -- --token YOUR_TOKEN
     ```

### Security Checklist:
- [ ] Build on clean CI/CD system
- [ ] Sign all packages
- [ ] Verify checksums
- [ ] Test on fresh VMs
- [ ] Review systemd hardening
- [ ] Check file permissions
- [ ] Validate enrollment process
- [ ] Test certificate rotation
- [ ] Verify auto-updates work

---

## Getting Help

- Documentation: `/home/tsho/ai-support-system/agent/docs/INSTALLATION.md`
- Packaging Guide: `/home/tsho/ai-support-system/agent/docs/PACKAGING.md`
- Verification Checklist: `/home/tsho/ai-support-system/agent/PACKAGING_VERIFICATION.md`

---

**Last Updated:** December 17, 2025
**Build Version:** 1.0.0 (Minimal Test Build)
**Status:** Ready for packaging testing, full functionality requires code fixes
