# JTNT Agent Installation Guide

Complete installation instructions for the JTNT RMM Agent across all supported platforms.

## Table of Contents

- [Quick Install](#quick-install)
- [Windows Installation](#windows-installation)
- [macOS Installation](#macos-installation)
- [Linux Installation](#linux-installation)
- [Enrollment](#enrollment)
- [Verification](#verification)
- [Upgrade](#upgrade)
- [Uninstallation](#uninstallation)
- [Troubleshooting](#troubleshooting)

## Quick Install

### Linux and macOS

The fastest way to install on Linux or macOS is using our universal installation script:

```bash
curl -fsSL https://install.jtnt.us/agent.sh | sudo bash -s -- --token YOUR_ENROLLMENT_TOKEN
```

This will automatically:
- Detect your operating system and architecture
- Download the appropriate package
- Install the agent
- Enroll with your hub
- Start the service

## Windows Installation

### Interactive Installation

1. Download the MSI installer:
   ```
   https://releases.jtnt.us/agent/latest/JTNT-Agent-x64.msi
   ```

2. Double-click the MSI file and follow the installation wizard

3. After installation, enroll the agent:
   ```cmd
   jtnt-agent enroll --token YOUR_TOKEN --hub https://hub.jtnt.us
   ```

### Silent Installation

For automated deployments or mass installations:

```cmd
msiexec /i JTNT-Agent-1.0.0-x64.msi /qn ^
  ENROLLMENT_TOKEN="your-enrollment-token" ^
  HUB_URL="https://hub.jtnt.us"
```

**Parameters:**
- `/i` - Install
- `/qn` - Silent mode (no UI)
- `ENROLLMENT_TOKEN` - Your enrollment token (optional, can enroll later)
- `HUB_URL` - Your hub URL (default: https://hub.jtnt.us)

### Silent Installation with Logging

To troubleshoot installation issues:

```cmd
msiexec /i JTNT-Agent-1.0.0-x64.msi /qn /l*v install.log ^
  ENROLLMENT_TOKEN="your-token"
```

View the log:
```cmd
notepad install.log
```

### Installation Locations

- **Binaries:** `C:\Program Files\JTNT\Agent\`
- **State & Data:** `C:\ProgramData\JTNT\Agent\`
- **Certificates:** `C:\ProgramData\JTNT\Agent\certs\`
- **Logs:** `C:\ProgramData\JTNT\Agent\logs\`

### Service Management

**Check service status:**
```cmd
sc query JTNTAgent
```

**Start service:**
```cmd
net start JTNTAgent
```

**Stop service:**
```cmd
net stop JTNTAgent
```

**Restart service:**
```cmd
net stop JTNTAgent && net start JTNTAgent
```

**View logs:**
```cmd
type "C:\ProgramData\JTNT\Agent\logs\agent.log"
```

## macOS Installation

### Interactive Installation

1. Download the PKG installer:
   ```bash
   curl -O https://releases.jtnt.us/agent/latest/JTNT-Agent.pkg
   ```

2. Install the package:
   ```bash
   sudo installer -pkg JTNT-Agent.pkg -target /
   ```

3. Enroll the agent:
   ```bash
   sudo jtnt-agent enroll --token YOUR_TOKEN --hub https://hub.jtnt.us
   ```

### Silent Installation with Enrollment

```bash
ENROLLMENT_TOKEN="your-token" HUB_URL="https://hub.jtnt.us" \
  sudo installer -pkg JTNT-Agent.pkg -target /
```

### Installation Locations

- **Binaries:** `/usr/local/jtnt/agent/`
- **State & Data:** `/Library/Application Support/JTNT/Agent/`
- **Certificates:** `/Library/Application Support/JTNT/Agent/certs/`
- **Logs:** `/Library/Application Support/JTNT/Agent/logs/`
- **Service Config:** `/Library/LaunchDaemons/us.jtnt.agentd.plist`

### Service Management

**Check service status:**
```bash
sudo launchctl list | grep jtnt
```

**Stop service:**
```bash
sudo launchctl unload -w /Library/LaunchDaemons/us.jtnt.agentd.plist
```

**Start service:**
```bash
sudo launchctl load -w /Library/LaunchDaemons/us.jtnt.agentd.plist
```

**Restart service:**
```bash
sudo launchctl unload /Library/LaunchDaemons/us.jtnt.agentd.plist
sudo launchctl load /Library/LaunchDaemons/us.jtnt.agentd.plist
```

**View logs:**
```bash
tail -f "/Library/Application Support/JTNT/Agent/logs/stdout.log"
```

## Linux Installation

### Debian/Ubuntu

#### Using APT (Recommended)

1. Download the DEB package:
   ```bash
   curl -O https://releases.jtnt.us/agent/latest/jtnt-agent_amd64.deb
   ```

2. Install:
   ```bash
   sudo apt install ./jtnt-agent_amd64.deb
   ```

3. Enroll:
   ```bash
   sudo jtnt-agent enroll --token YOUR_TOKEN --hub https://hub.jtnt.us
   ```

4. Start service:
   ```bash
   sudo systemctl start jtnt-agentd
   ```

#### Using dpkg

```bash
sudo dpkg -i jtnt-agent_amd64.deb
sudo apt-get install -f  # Fix any dependency issues
```

### Installation Locations

- **Binaries:** `/usr/local/bin/`
- **State & Data:** `/var/lib/jtnt-agent/`
- **Configuration:** `/etc/jtnt-agent/`
- **Certificates:** `/var/lib/jtnt-agent/certs/`
- **Logs:** `/var/lib/jtnt-agent/logs/`
- **Service Unit:** `/lib/systemd/system/jtnt-agentd.service`

### Service Management

**Check service status:**
```bash
sudo systemctl status jtnt-agentd
```

**Start service:**
```bash
sudo systemctl start jtnt-agentd
```

**Stop service:**
```bash
sudo systemctl stop jtnt-agentd
```

**Restart service:**
```bash
sudo systemctl restart jtnt-agentd
```

**Enable on boot:**
```bash
sudo systemctl enable jtnt-agentd
```

**View logs:**
```bash
sudo journalctl -u jtnt-agentd -f
```

## Enrollment

After installation, enroll the agent with your JTNT Hub:

### Interactive Enrollment

```bash
sudo jtnt-agent enroll --token YOUR_ENROLLMENT_TOKEN
```

The hub URL defaults to `https://hub.jtnt.us`. To use a different hub:

```bash
sudo jtnt-agent enroll --token YOUR_TOKEN --hub https://your-hub.example.com
```

### Automated Enrollment

For scripted deployments, pass the token during installation:

**Windows:**
```cmd
msiexec /i JTNT-Agent.msi /qn ENROLLMENT_TOKEN="your-token"
```

**macOS:**
```bash
ENROLLMENT_TOKEN="your-token" sudo installer -pkg JTNT-Agent.pkg -target /
```

**Linux:**
```bash
curl -fsSL https://install.jtnt.us/agent.sh | sudo bash -s -- --token YOUR_TOKEN
```

### Verify Enrollment

Check that the agent has successfully enrolled:

```bash
sudo jtnt-agent status
```

Expected output:
```
Agent Status: Enrolled
Agent ID: 550e8400-e29b-41d4-a716-446655440000
Hub: https://hub.jtnt.us
Last Heartbeat: 2025-12-16 10:30:00 UTC
Certificate Expires: 2026-12-16 00:00:00 UTC
```

## Verification

### Check Service is Running

**Windows:**
```cmd
sc query JTNTAgent
```
Should show `STATE: RUNNING`

**macOS:**
```bash
sudo launchctl list | grep jtnt
```
Should show the agent in the list

**Linux:**
```bash
sudo systemctl is-active jtnt-agentd
```
Should output `active`

### Check Network Connectivity

Verify the agent can reach the hub:

```bash
sudo jtnt-agent heartbeat
```

Expected output:
```
Heartbeat sent successfully
Response: OK
```

### View Metrics

Check Prometheus metrics (localhost only):

```bash
curl http://localhost:9090/metrics
```

### View Health Status

Check health endpoint (localhost only):

```bash
curl http://localhost:9091/health
```

## Upgrade

### In-Place Upgrade

The agent supports in-place upgrades that preserve your agent ID, certificates, and configuration.

**Windows:**
```cmd
msiexec /i JTNT-Agent-1.1.0-x64.msi /qn
```

**macOS:**
```bash
sudo installer -pkg JTNT-Agent-1.1.0.pkg -target /
```

**Linux:**
```bash
sudo apt install ./jtnt-agent_1.1.0_amd64.deb
# or
sudo dpkg -i jtnt-agent_1.1.0_amd64.deb
```

### Automatic Updates

The agent can automatically update itself when new versions are available. This is controlled by hub policy.

To check for updates manually:

```bash
sudo jtnt-agent update check
```

To apply an available update:

```bash
sudo jtnt-agent update apply
```

### Rollback

If an update fails, the agent automatically rolls back to the previous version. You can also manually rollback:

```bash
sudo jtnt-agent update rollback
```

## Uninstallation

### Windows

**Interactive:**
```cmd
msiexec /x JTNT-Agent-1.0.0-x64.msi
```

**Silent:**
```cmd
msiexec /x JTNT-Agent-1.0.0-x64.msi /qn
```

**Remove by product code:**
```cmd
wmic product where name="JTNT Agent" call uninstall
```

### macOS

```bash
sudo /usr/local/jtnt/agent/uninstall.sh
```

Or use the universal uninstall script:

```bash
curl -fsSL https://install.jtnt.us/uninstall.sh | sudo bash
```

To remove all data:
```bash
sudo /usr/local/jtnt/agent/uninstall.sh --purge
```

### Linux

**Keep data:**
```bash
sudo apt-get remove jtnt-agent
# or
sudo dpkg --remove jtnt-agent
```

**Remove all data:**
```bash
sudo apt-get purge jtnt-agent
# or
sudo dpkg --purge jtnt-agent
```

**Universal uninstall script:**
```bash
curl -fsSL https://install.jtnt.us/uninstall.sh | sudo bash --purge
```

## Troubleshooting

### Installation Fails

**Windows:**
- Check the installation log: `msiexec /i JTNT-Agent.msi /l*v install.log`
- Verify you have Administrator privileges
- Check Windows version (requires Windows 10 or Server 2016+)

**macOS:**
- Check installer log: `cat /var/log/install.log | grep -i jtnt`
- Verify macOS version (requires macOS 13.0+)
- Check System Preferences > Security & Privacy for blocked installers

**Linux:**
- Check package installation: `sudo dpkg -l | grep jtnt`
- Fix dependencies: `sudo apt-get install -f`
- Check systemd status: `sudo systemctl status jtnt-agentd`

### Service Won't Start

**Check logs:**
- Windows: `C:\ProgramData\JTNT\Agent\logs\`
- macOS: `/Library/Application Support/JTNT/Agent/logs/`
- Linux: `sudo journalctl -u jtnt-agentd -n 50`

**Common issues:**
- Missing enrollment: Run `sudo jtnt-agent enroll --token YOUR_TOKEN`
- Invalid certificates: Re-enroll the agent
- Port conflicts: Check if ports 9090/9091 are available
- Permissions: Verify state directory permissions

### Enrollment Fails

**Check:**
- Token is valid and not expired
- Hub URL is correct and accessible
- Network connectivity to hub
- Firewall allows outbound HTTPS (443)

**Debug enrollment:**
```bash
sudo jtnt-agent enroll --token YOUR_TOKEN --debug
```

### Agent Not Communicating

**Verify connectivity:**
```bash
curl -v https://hub.jtnt.us/health
```

**Check firewall:**
- Ensure outbound HTTPS (443) is allowed
- Check for proxy requirements

**Test heartbeat:**
```bash
sudo jtnt-agent heartbeat --debug
```

### High CPU or Memory Usage

**Check resource limits:**
```bash
# Linux
sudo systemctl show jtnt-agentd | grep -E "(CPU|Memory)"

# macOS
ps aux | grep jtnt-agentd
```

**View metrics:**
```bash
curl http://localhost:9090/metrics | grep process
```

### Certificate Errors

**Check certificate expiration:**
```bash
sudo jtnt-agent cert info
```

**Force certificate renewal:**
```bash
sudo jtnt-agent cert renew
```

**Re-enroll (last resort):**
```bash
sudo jtnt-agent unenroll
sudo jtnt-agent enroll --token NEW_TOKEN
```

## Getting Help

- **Documentation:** https://docs.jtnt.us/agent
- **Support Email:** support@jtnt.us
- **GitHub Issues:** https://github.com/tshojoshua/ai-support-system-agent/issues
- **Community Forum:** https://community.jtnt.us

## System Requirements

### Windows
- Windows 10 or Windows Server 2016 or later
- 50 MB free disk space
- Network connectivity to JTNT Hub

### macOS
- macOS 13.0 (Ventura) or later
- Intel or Apple Silicon processor
- 50 MB free disk space
- Network connectivity to JTNT Hub

### Linux
- Ubuntu 20.04+, Debian 11+, or compatible
- systemd-based distribution
- 50 MB free disk space
- Network connectivity to JTNT Hub
