# JTNT Agent Packaging Guide

Complete guide for building, signing, and distributing JTNT Agent installers across all supported platforms.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Building Packages](#building-packages)
- [Signing and Notarization](#signing-and-notarization)
- [CI/CD Pipeline](#cicd-pipeline)
- [Testing Packages](#testing-packages)
- [Distribution](#distribution)

## Overview

The JTNT Agent supports three packaging formats:

| Platform | Format | Tool | Output |
|----------|--------|------|--------|
| Windows | MSI | WiX Toolset | `JTNT-Agent-{version}-x64.msi` |
| macOS | PKG | pkgbuild | `JTNT-Agent-{version}.pkg` |
| Linux | DEB | dpkg-deb | `jtnt-agent_{version}_{arch}.deb` |

All packages include:
- Agent daemon binary
- CLI tool
- Service configuration
- Automatic enrollment support
- Upgrade-in-place capability

## Prerequisites

### All Platforms
- Go 1.23 or later
- Git
- Code signing certificates (for production builds)

### Windows
- Windows 10 or later
- [WiX Toolset](https://wixtoolset.org/) 3.11+ or 4.0+
- PowerShell 5.1 or PowerShell Core 7+
- Code signing certificate (optional, for signing MSI)

### macOS
- macOS 13.0 or later (for building)
- Xcode Command Line Tools
- Developer ID certificate (for signing)
- Apple Developer account (for notarization)

### Linux
- Ubuntu 20.04+ or Debian 11+ (recommended for building)
- dpkg-dev
- debhelper
- lintian (optional, for validation)

## Building Packages

### Windows MSI

#### Quick Build

```powershell
cd packaging/windows
.\build.ps1
```

#### Build Specific Version

```powershell
.\build.ps1 -Version "1.2.3" -Platform "x64"
```

#### Build Process

The build script:
1. Validates WiX Toolset is installed
2. Builds Go binaries (jtnt-agentd.exe, jtnt-agent.exe)
3. Compiles WiX source files (.wxs → .wixobj)
4. Links WiX objects into MSI
5. Outputs to `packaging/windows/output/`

#### Customization

**Product.wxs:**
- Product GUID (UpgradeCode) - Must remain constant across versions
- Component GUIDs - Generate new for each new component
- Installation paths
- Service configuration
- Custom actions

**Generate new GUIDs:**
```powershell
[guid]::NewGuid().ToString().ToUpper()
```

#### Output

- MSI installer: `JTNT-Agent-{version}-x64.msi`
- Size: ~8-12 MB (compressed)

### macOS PKG

#### Quick Build

```bash
cd packaging/macos
chmod +x build.sh
./build.sh
```

#### Build Specific Version

```bash
./build.sh 1.2.3
```

#### Build Process

The build script:
1. Builds universal binaries (Intel + Apple Silicon)
   - Compiles for amd64 and arm64 separately
   - Uses `lipo` to create universal binary
2. Creates package structure in `pkg-root/`
3. Copies binaries and launchd plist
4. Builds component package with scripts
5. Creates distribution package with `productbuild`
6. Signs package (if SIGNING_IDENTITY set)
7. Outputs to `packaging/macos/output/`

#### Universal Binary Verification

```bash
lipo -info bin/jtnt-agentd
# Expected: Architectures in the fat file: bin/jtnt-agentd are: x86_64 arm64
```

#### Customization

**Distribution.xml:**
- Installer UI customization
- Installation checks
- Welcome/Conclusion messages

**Scripts:**
- `preinstall` - Runs before installation
- `postinstall` - Runs after installation, sets up service
- Customize enrollment, permissions, etc.

#### Output

- PKG installer: `JTNT-Agent-{version}.pkg`
- Size: ~16-20 MB (universal binary)

### Linux DEB

#### Quick Build

```bash
cd packaging/linux
chmod +x build.sh
./build.sh
```

#### Build Specific Version and Architecture

```bash
./build.sh 1.2.3 amd64
./build.sh 1.2.3 arm64
```

#### Build Process

The build script:
1. Builds Linux binary for specified architecture
2. Creates Debian package structure in `deb-root/`
3. Copies binaries and systemd unit
4. Generates control file from template
5. Copies maintainer scripts (postinst, prerm, postrm)
6. Creates copyright and changelog
7. Builds package with `dpkg-deb`
8. Validates with `lintian` (if available)
9. Outputs to `packaging/linux/output/`

#### Customization

**control.template:**
- Package metadata (name, version, dependencies)
- Architecture
- Description

**Maintainer Scripts:**
- `postinst` - Creates user, sets permissions, enables service
- `prerm` - Stops service before removal
- `postrm` - Cleanup after removal, purge handling

**systemd Unit:**
- Service configuration
- Security hardening options
- Resource limits

#### Supported Architectures

- `amd64` - 64-bit Intel/AMD
- `arm64` - 64-bit ARM (Raspberry Pi 4, AWS Graviton)
- `armhf` - 32-bit ARM (older Raspberry Pi)

#### Output

- DEB package: `jtnt-agent_{version}_{arch}.deb`
- Size: ~8-12 MB per architecture

## Signing and Notarization

### Windows Code Signing

#### With signtool

```powershell
# Sign MSI
signtool sign /f certificate.pfx /p password /tr http://timestamp.digicert.com `
  /td SHA256 /fd SHA256 JTNT-Agent-1.0.0-x64.msi

# Verify signature
signtool verify /pa JTNT-Agent-1.0.0-x64.msi
```

#### With Azure Key Vault

```powershell
# Sign using Azure Key Vault certificate
AzureSignTool sign -kvu "https://your-vault.vault.azure.net" `
  -kvi "client-id" -kvs "client-secret" -kvc "cert-name" `
  -tr http://timestamp.digicert.com JTNT-Agent-1.0.0-x64.msi
```

### macOS Signing and Notarization

#### Sign Package

```bash
# Set environment variable
export SIGNING_IDENTITY="Developer ID Installer: Your Name (TEAM_ID)"

# Build with signing
cd packaging/macos
./build.sh 1.0.0

# Manual signing
productsign --sign "$SIGNING_IDENTITY" \
  JTNT-Agent-1.0.0.pkg \
  JTNT-Agent-1.0.0-signed.pkg
```

#### Notarize with Apple

```bash
# Upload for notarization
xcrun notarytool submit JTNT-Agent-1.0.0-signed.pkg \
  --apple-id "your-email@example.com" \
  --team-id "TEAM_ID" \
  --password "app-specific-password" \
  --wait

# Check notarization status
xcrun notarytool info <submission-id> \
  --apple-id "your-email@example.com" \
  --team-id "TEAM_ID" \
  --password "app-specific-password"

# Staple notarization ticket
xcrun stapler staple JTNT-Agent-1.0.0-signed.pkg

# Verify
xcrun stapler validate JTNT-Agent-1.0.0-signed.pkg
spctl -a -v --type install JTNT-Agent-1.0.0-signed.pkg
```

### Linux Package Signing

#### Sign DEB with GPG

```bash
# Create detached signature
gpg --armor --detach-sign jtnt-agent_1.0.0_amd64.deb

# Verify signature
gpg --verify jtnt-agent_1.0.0_amd64.deb.asc jtnt-agent_1.0.0_amd64.deb
```

#### Create APT Repository

```bash
# Create repository structure
mkdir -p repo/dists/stable/main/binary-amd64

# Copy packages
cp *.deb repo/dists/stable/main/binary-amd64/

# Generate Packages file
cd repo
dpkg-scanpackages dists/stable/main/binary-amd64 /dev/null | \
  gzip -9c > dists/stable/main/binary-amd64/Packages.gz

# Create Release file
cat > dists/stable/Release <<EOF
Origin: JTNT
Label: JTNT Agent Repository
Suite: stable
Codename: stable
Architectures: amd64 arm64
Components: main
Description: JTNT Agent APT Repository
EOF

# Sign Release
gpg --clearsign -o dists/stable/InRelease dists/stable/Release
```

### Ed25519 Signature (All Platforms)

For update verification, sign all packages with Ed25519:

```bash
# Generate key pair (one time)
openssl genpkey -algorithm ED25519 -out private-key.pem
openssl pkey -in private-key.pem -pubout -out public-key.pem

# Sign package
openssl pkeyutl -sign -inkey private-key.pem \
  -rawin -in <(sha256sum package.ext | cut -d' ' -f1 | xxd -r -p) \
  -out package.ext.sig

# Verify (agent does this)
openssl pkeyutl -verify -pubin -inkey public-key.pem \
  -rawin -in <(sha256sum package.ext | cut -d' ' -f1 | xxd -r -p) \
  -sigfile package.ext.sig
```

## CI/CD Pipeline

### GitHub Actions Workflow

The repository includes a complete CI/CD workflow at `.github/workflows/build.yml`.

#### Trigger Release Build

**Manual trigger:**
```bash
# From GitHub UI: Actions → Build and Release → Run workflow → Enter version

# Or using gh CLI:
gh workflow run build.yml -f version=1.2.3
```

**Tag-based trigger:**
```bash
git tag v1.2.3
git push origin v1.2.3
```

#### Workflow Steps

1. **Build Windows MSI** (windows-latest runner)
   - Setup Go
   - Install WiX Toolset
   - Build MSI
   - Upload artifact

2. **Build macOS PKG** (macos-latest runner)
   - Setup Go
   - Build universal binaries
   - Create PKG
   - Upload artifact

3. **Build Linux DEB** (ubuntu-latest runner, matrix: amd64, arm64)
   - Setup Go
   - Install dpkg tools
   - Build DEB for each architecture
   - Upload artifacts

4. **Sign and Release** (ubuntu-latest runner)
   - Download all artifacts
   - Sign with Ed25519 key (from secrets)
   - Generate SHA256 checksums
   - Create GitHub release
   - Upload all files

5. **Notify** (ubuntu-latest runner)
   - Send Slack notification (if configured)

#### Required Secrets

Configure in GitHub repository settings (Settings → Secrets and variables → Actions):

| Secret | Description |
|--------|-------------|
| `SIGNING_KEY` | Base64-encoded Ed25519 private key for signing packages |
| `SLACK_WEBHOOK_URL` | (Optional) Slack webhook for release notifications |

#### Generate Signing Key

```bash
# Generate Ed25519 key pair
openssl genpkey -algorithm ED25519 -outform DER -out signing-key.der
openssl pkey -in signing-key.der -inform DER -pubout -outform DER -out public-key.der

# Base64 encode for GitHub secret
base64 < signing-key.der > signing-key.b64

# Add signing-key.b64 content to SIGNING_KEY secret
# Embed public-key.der in agent binary for verification
```

### Local Build Script

For building all platforms locally (requires each platform's tools):

```bash
#!/bin/bash
VERSION=${1:-1.0.0}

# Build Windows (requires Windows or Wine)
cd packaging/windows && ./build.ps1 -Version $VERSION

# Build macOS (requires macOS)
cd ../macos && ./build.sh $VERSION

# Build Linux DEB
cd ../linux && ./build.sh $VERSION amd64
./build.sh $VERSION arm64

echo "All packages built for version $VERSION"
```

## Testing Packages

### Automated Testing

Create a test matrix:

```yaml
# .github/workflows/test-packages.yml
name: Test Packages

on:
  workflow_run:
    workflows: ["Build and Release Agent"]
    types: [completed]

jobs:
  test-windows:
    runs-on: windows-latest
    steps:
      - name: Download MSI
        # ... download artifact
      - name: Install silently
        run: msiexec /i JTNT-Agent.msi /qn
      - name: Verify service
        run: sc query JTNTAgent
      - name: Uninstall
        run: msiexec /x JTNT-Agent.msi /qn

  test-macos:
    runs-on: macos-latest
    steps:
      - name: Download PKG
        # ... download artifact
      - name: Install
        run: sudo installer -pkg JTNT-Agent.pkg -target /
      - name: Verify service
        run: launchctl list | grep jtnt
      - name: Uninstall
        run: sudo /usr/local/jtnt/agent/uninstall.sh --force

  test-linux:
    runs-on: ubuntu-latest
    steps:
      - name: Download DEB
        # ... download artifact
      - name: Install
        run: sudo dpkg -i jtnt-agent.deb
      - name: Verify service
        run: systemctl is-active jtnt-agentd
      - name: Uninstall
        run: sudo dpkg --purge jtnt-agent
```

### Manual Testing Checklist

#### Fresh Install
- [ ] Package installs without errors
- [ ] Service starts automatically
- [ ] Enrollment works with token
- [ ] Agent connects to hub
- [ ] Metrics endpoint accessible (localhost)
- [ ] Health endpoint accessible (localhost)

#### Upgrade
- [ ] Upgrade from previous version succeeds
- [ ] Agent ID preserved
- [ ] Certificates preserved
- [ ] Configuration preserved
- [ ] Service restarts automatically

#### Uninstall
- [ ] Service stops cleanly
- [ ] Binaries removed
- [ ] Service configuration removed
- [ ] Data preserved (default)
- [ ] Purge removes all data

#### Platform-Specific

**Windows:**
- [ ] PATH updated correctly
- [ ] Service runs as NetworkService
- [ ] Event log entries created
- [ ] MSI upgrade/downgrade protection works

**macOS:**
- [ ] Universal binary works on Intel
- [ ] Universal binary works on Apple Silicon
- [ ] launchd plist valid
- [ ] Permissions correct on state directory

**Linux:**
- [ ] systemd unit starts correctly
- [ ] User/group created
- [ ] Security hardening options work
- [ ] journald logging works

## Distribution

### Release Hosting

Packages should be hosted at: `https://releases.jtnt.us/agent/{version}/`

Directory structure:
```
releases/
├── latest.txt              # Contains latest version number
├── 1.0.0/
│   ├── JTNT-Agent-1.0.0-x64.msi
│   ├── JTNT-Agent-1.0.0-x64.msi.sha256
│   ├── JTNT-Agent-1.0.0-x64.msi.sig
│   ├── JTNT-Agent-1.0.0.pkg
│   ├── JTNT-Agent-1.0.0.pkg.sha256
│   ├── JTNT-Agent-1.0.0.pkg.sig
│   ├── jtnt-agent_1.0.0_amd64.deb
│   ├── jtnt-agent_1.0.0_amd64.deb.sha256
│   ├── jtnt-agent_1.0.0_amd64.deb.sig
│   ├── jtnt-agent_1.0.0_arm64.deb
│   ├── jtnt-agent_1.0.0_arm64.deb.sha256
│   └── jtnt-agent_1.0.0_arm64.deb.sig
└── 1.1.0/
    └── ...
```

### Update latest.txt

```bash
echo "1.2.3" > latest.txt
# Upload to releases.jtnt.us/agent/latest.txt
```

### CDN Configuration

Use CloudFront or similar CDN for global distribution:

- Cache-Control headers for packages (immutable, 1 year)
- Short cache for latest.txt (5 minutes)
- HTTPS only
- Signed URLs for authenticated access (optional)

### Verification Instructions

Include in release notes:

```markdown
## Verification

All packages are signed with Ed25519 and SHA256 checksums provided.

**Verify SHA256:**
```bash
# Windows
certutil -hashfile JTNT-Agent-1.0.0-x64.msi SHA256

# macOS/Linux
sha256sum JTNT-Agent-1.0.0.pkg
```

**Verify Signature:**
Packages are signed with our Ed25519 public key (embedded in agent).
The agent verifies signatures automatically during updates.
```

## Troubleshooting

### Build Failures

**WiX not found (Windows):**
```powershell
# Check WiX installation
candle.exe -?

# Add to PATH
$env:Path += ";C:\Program Files (x86)\WiX Toolset v3.11\bin"
```

**lipo fails (macOS):**
- Ensure both amd64 and arm64 binaries exist
- Check binary format: `file bin/jtnt-agentd-amd64`

**dpkg-deb fails (Linux):**
```bash
# Check package structure
dpkg-deb --contents deb-root/

# Fix permissions
chmod 755 deb-root/DEBIAN/postinst
```

### Signing Issues

**Certificate not found:**
- Windows: Import PFX to certificate store
- macOS: Check with `security find-identity -v -p codesigning`

**Notarization fails:**
- Check app-specific password
- Verify TEAM_ID correct
- Review notarization log: `xcrun notarytool log <id>`

## Best Practices

1. **Version Consistency:** Use semantic versioning (MAJOR.MINOR.PATCH)
2. **Changelog:** Update CHANGELOG.md before each release
3. **Testing:** Test on minimum supported OS versions
4. **Signing:** Always sign production packages
5. **Backup:** Keep previous versions available for rollback
6. **Documentation:** Update docs before releasing
7. **Communication:** Announce releases to users

## Resources

- [WiX Toolset Documentation](https://wixtoolset.org/documentation/)
- [Apple PKG Documentation](https://developer.apple.com/library/archive/documentation/DeveloperTools/Reference/DistributionDefinitionRef/)
- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [systemd Unit Documentation](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
