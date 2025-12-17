# Windows MSI Packaging

This directory contains the WiX source files for building the JTNT Agent MSI installer for Windows.

## Prerequisites

- WiX Toolset 3.11+ or WiX 4.0+ (https://wixtoolset.org/)
- Go 1.23+ (https://go.dev/)
- PowerShell 5.1+ or PowerShell Core 7+

## Building the MSI

### Quick Build

```powershell
.\build.ps1
```

### Build with Specific Version

```powershell
.\build.ps1 -Version "1.2.3"
```

### Build for Specific Platform

```powershell
.\build.ps1 -Version "1.0.0" -Platform "x64"
```

## Installation

### Interactive Installation

```cmd
msiexec /i JTNT-Agent-1.0.0-x64.msi
```

### Silent Installation

```cmd
msiexec /i JTNT-Agent-1.0.0-x64.msi /qn
```

### Silent Installation with Enrollment

```cmd
msiexec /i JTNT-Agent-1.0.0-x64.msi /qn ^
  ENROLLMENT_TOKEN="your-enrollment-token" ^
  HUB_URL="https://hub.jtnt.us"
```

### Installation with Logging

```cmd
msiexec /i JTNT-Agent-1.0.0-x64.msi /l*v install.log
```

## Uninstallation

### Interactive Uninstallation

```cmd
msiexec /x JTNT-Agent-1.0.0-x64.msi
```

### Silent Uninstallation

```cmd
msiexec /x JTNT-Agent-1.0.0-x64.msi /qn
```

### Uninstall by Product Code

```cmd
msiexec /x {PRODUCT-GUID} /qn
```

## Upgrade

The MSI supports in-place upgrades. Simply install a newer version over an existing installation:

```cmd
msiexec /i JTNT-Agent-1.1.0-x64.msi /qn
```

**Note:** Agent ID and certificates are preserved during upgrades.

## Files

- `Product.wxs` - Main WiX product definition
- `build.ps1` - PowerShell build script
- `en-us.wxl` - English localization strings
- `license.rtf` - End-user license agreement
- `banner.bmp` - Installer banner (493x58 pixels) - PLACEHOLDER

## Directory Structure

### Installation Directory
- Default: `C:\Program Files\JTNT\Agent\`
- Contains: `jtnt-agentd.exe`, `jtnt-agent.exe`

### State Directory
- Location: `C:\ProgramData\JTNT\Agent\`
- Contains: `certs\`, `logs\`, configuration files
- **Preserved during upgrades**

## Service Configuration

- **Service Name:** JTNTAgent
- **Display Name:** JTNT Agent
- **Description:** JTNT RMM Agent for secure remote management and monitoring
- **Account:** NT AUTHORITY\NetworkService
- **Start Type:** Automatic

## MSI Properties

| Property | Default | Description |
|----------|---------|-------------|
| ENROLLMENT_TOKEN | (empty) | Enrollment token for automatic enrollment |
| HUB_URL | https://hub.jtnt.us | Hub URL for enrollment |
| INSTALLFOLDER | C:\Program Files\JTNT\Agent | Installation directory |

## Return Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1602 | User cancelled |
| 1603 | Fatal error during installation |
| 1638 | Another version already installed |
| 3010 | Reboot required |

## Troubleshooting

### View MSI Log

```cmd
msiexec /i JTNT-Agent-1.0.0-x64.msi /l*v install.log
notepad install.log
```

### Check Service Status

```powershell
Get-Service JTNTAgent
```

### View Event Logs

```powershell
Get-EventLog -LogName Application -Source "MsiInstaller" -Newest 20
```

### Manual Service Start

```cmd
net start JTNTAgent
```

## Development

### Generate New GUIDs

```powershell
[guid]::NewGuid().ToString().ToUpper()
```

### Validate WXS File

```cmd
candle.exe -nologo Product.wxs
```

### Custom Banner

Replace `banner.bmp` with a 493x58 pixel, 24-bit BMP file.

## Security

- Service runs as NetworkService (low privilege)
- State directory permissions restricted
- Certificates stored with restricted ACLs
- Automatic updates signed and verified
