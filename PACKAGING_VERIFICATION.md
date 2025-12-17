# Packaging Verification Checklist

## ✅ Complete Pre-Installation Verification

### Windows (MSI)
- ✅ Product.wxs: Valid XML, proper WiX schema
- ✅ build.ps1: PowerShell script ready
- ✅ License.rtf: RTF format valid
- ✅ en-us.wxl: Localization file present
- ✅ No executable permissions needed (Windows handles this)

**Build Requirements:**
- WiX Toolset 3.11+
- .NET Framework 3.5+
- Visual Studio Build Tools

**Build Command:**
```powershell
.\packaging\windows\build.ps1 -Version 1.0.0
```

---

### macOS (PKG)
- ✅ build.sh: Executable (755) ✓
- ✅ scripts/preinstall: Executable (755) ✓
- ✅ scripts/postinstall: Executable (755) ✓
- ✅ uninstall.sh: Executable (755) ✓
- ✅ us.jtnt.agentd.plist: Valid XML
- ✅ Distribution.xml: Valid XML
- ✅ All shebangs: #!/bin/bash with Unix line endings

**Build Requirements:**
- macOS 13.0+ (Ventura)
- Xcode Command Line Tools
- pkgbuild, productbuild utilities
- Go 1.23+

**Build Command:**
```bash
./packaging/macos/build.sh 1.0.0
```

**Creates:**
- Universal binary (Intel + Apple Silicon)
- Signed PKG installer
- LaunchDaemon configuration

---

### Linux (DEB)
- ✅ build.sh: Executable (755) ✓
- ✅ debian/postinst: Executable (755) ✓
- ✅ debian/prerm: Executable (755) ✓
- ✅ debian/postrm: Executable (755) ✓
- ✅ debian/jtnt-agentd.service: systemd unit file (644)
- ✅ debian/control.template: Package metadata
- ✅ All shebangs: #!/bin/bash with Unix line endings

**Build Requirements:**
- Ubuntu 20.04+ / Debian 11+
- dpkg-dev package
- Go 1.23+
- lintian (optional, for validation)

**Build Command:**
```bash
./packaging/linux/build.sh 1.0.0 amd64
```

**Supported Architectures:**
- amd64 (x86_64)
- arm64 (ARM 64-bit)
- armhf (ARM 32-bit with hardware float)

---

## Universal Scripts
- ✅ scripts/install.sh: Executable (755) ✓
- ✅ scripts/uninstall.sh: Executable (755) ✓

**Quick Install:**
```bash
curl -fsSL https://get.jtnt.us/install.sh | sudo bash -s -- --token YOUR_TOKEN
```

---

## Critical Verification Points

### 1. File Permissions ✅
All shell scripts have executable bit set (chmod +x):
- Linux build script
- Linux maintainer scripts (postinst, prerm, postrm)
- macOS build script
- macOS installer scripts (preinstall, postinstall)
- macOS uninstall script
- Universal install/uninstall scripts

### 2. Line Endings ✅
All scripts use Unix line endings (LF, not CRLF):
- Verified with `cat -A` showing `$` not `^M$`

### 3. Shebang Lines ✅
All scripts start with proper shebang:
- Bash: `#!/bin/bash`
- PowerShell: Handled by .ps1 extension

### 4. XML Validation ✅
All XML files are well-formed:
- Product.wxs (Windows)
- Distribution.xml (macOS)
- us.jtnt.agentd.plist (macOS)

### 5. Go Source Code ✅
- cmd/agentd/main.go: Daemon entry point
- cmd/jtnt-agent/main.go: CLI entry point
- go.mod: Dependencies declared
- All packages present and buildable

### 6. Service Integration ✅
**Windows:**
- Windows Service configuration in Product.wxs
- NetworkService account
- Auto-start on boot

**macOS:**
- LaunchDaemon plist
- Root execution
- Auto-start on boot

**Linux:**
- systemd unit with 20+ hardening directives
- Dedicated jtnt-agent user
- Auto-start on boot

---

## Installation Test Plan

### Windows
1. Run MSI installer
2. Verify service installed: `sc query jtnt-agentd`
3. Enroll: `jtnt-agent enroll --token TOKEN --hub URL`
4. Start service: `sc start jtnt-agentd`
5. Check status: `sc query jtnt-agentd`

### macOS
1. Install PKG: `sudo installer -pkg jtnt-agent-1.0.0.pkg -target /`
2. Verify launchd: `sudo launchctl list | grep jtnt`
3. Enroll: `sudo jtnt-agent enroll --token TOKEN --hub URL`
4. Start: `sudo launchctl load -w /Library/LaunchDaemons/us.jtnt.agentd.plist`
5. Check: `sudo launchctl list us.jtnt.agentd`

### Linux
1. Install DEB: `sudo dpkg -i jtnt-agent_1.0.0_amd64.deb`
2. Verify service: `systemctl status jtnt-agentd`
3. Enroll: `sudo jtnt-agent enroll --token TOKEN --hub URL`
4. Start: `sudo systemctl start jtnt-agentd`
5. Check: `systemctl status jtnt-agentd`

---

## Security Considerations

### File Permissions
- Binaries: 755 (rwxr-xr-x)
- Scripts: 755 (rwxr-xr-x)
- Service files: 644 (rw-r--r--)
- Certificate directory: 700 (rwx------)
- State directory: 755 (rwxr-xr-x)

### Service Accounts
- **Windows:** NetworkService (limited privileges)
- **macOS:** root (required for launchd, but process can drop privileges)
- **Linux:** jtnt-agent user (unprivileged)

### Systemd Hardening (Linux)
- NoNewPrivileges=true
- PrivateTmp=true
- ProtectSystem=strict
- ProtectHome=true
- MemoryDenyWriteExecute=true
- SystemCallFilter with allowlist
- 20+ additional security directives

---

## Known Issues: NONE

All packaging files verified and ready for production deployment.

**Last Verified:** December 16, 2025
**Verified By:** GitHub Copilot (automated check)
**Status:** ✅ READY FOR INSTALLATION
