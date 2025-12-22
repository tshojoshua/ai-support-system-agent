# VS Code Setup on portal.tsho.us for MSI Build

## Required VS Code Extensions

Install these extensions on portal.tsho.us Windows machine:

```powershell
# In PowerShell on Windows:
# Install VS Code extensions via command line

# 1. Go extension (CRITICAL)
code --install-extension golang.go

# 2. GitHub Copilot (AI assistance)
code --install-extension GitHub.copilot
code --install-extension GitHub.copilot-chat

# 3. PowerShell extension (for build scripts)
code --install-extension ms-vscode.PowerShell

# 4. XML extension (for WiX Product.wxs)
code --install-extension redhat.vscode-xml

# 5. Markdown (for documentation)
code --install-extension yzhang.markdown-all-in-one
```

### Manual Install (if command line doesn't work)
1. Press `Ctrl+Shift+X` in VS Code
2. Search for each extension:
   - **Go** (golang.go) - REQUIRED
   - **GitHub Copilot** (GitHub.copilot) - REQUIRED
   - **GitHub Copilot Chat** (GitHub.copilot-chat) - REQUIRED
   - **PowerShell** (ms-vscode.PowerShell)
   - **XML** (redhat.vscode-xml)
   - **Markdown All in One** (yzhang.markdown-all-in-one)
3. Click "Install" for each

## Go Extension Setup

After installing Go extension, VS Code will prompt to install Go tools:

1. Press `Ctrl+Shift+P`
2. Type: `Go: Install/Update Tools`
3. Select ALL tools, click OK
4. Wait for installation (~2-3 minutes)

Required Go tools:
- `gopls` (language server)
- `go-outline`
- `gotests`
- `gomodifytags`
- `impl`
- `dlv` (debugger)

## Context for Copilot Agent

When you open the project in VS Code on Windows, give Copilot this context in chat:

```
I'm building a Windows MSI installer for JTNT RMM Agent v4.0.0.

Project structure:
- Go 1.23+ application
- Main binaries: cmd/jtnt-agent/main.go and cmd/agentd/main.go
- WiX MSI config: packaging/windows/Product.wxs
- Build script: packaging/windows/build.ps1
- Dependencies in go.mod

Task: Build MSI installer using WiX Toolset 3.11+

Current build command:
```powershell
cd packaging\windows
.\build.ps1 -Version "4.0.0"
```

Requirements:
1. Compile Go binaries for Windows (amd64)
2. Package with WiX into MSI
3. MSI should install Windows service (jtnt-agentd)
4. Auto-enroll agent with token
5. Output: JTNT-Agent-4.0.0-x64.msi

Build environment:
- Windows 10/11 or Server 2019/2022
- Go 1.23+ installed via Chocolatey
- WiX Toolset 3.11+ installed via Chocolatey
- PowerShell 5.1+

Key files to reference:
- packaging/windows/BUILD_GUIDE.md (comprehensive instructions)
- packaging/windows/Product.wxs (WiX configuration)
- packaging/windows/build.ps1 (build automation)
- EMERGENCY_DEPLOYMENT.md (deployment timeline)
- GO-NOGO.md (critical blockers to check)

Current blockers from GO-NOGO.md:
1. Windows MSI missing (this is what we're building)
2. Hub API endpoints returning 404 (need verification)

Timeline: Emergency deployment tonight, need MSI built ASAP.
```

## VS Code Tasks Configuration

Create `.vscode/tasks.json` on Windows portal:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build Go Binaries",
      "type": "shell",
      "command": "go",
      "args": [
        "build",
        "-o",
        "bin/jtnt-agent.exe",
        "./cmd/jtnt-agent"
      ],
      "group": "build",
      "presentation": {
        "reveal": "always",
        "panel": "new"
      },
      "problemMatcher": ["$go"]
    },
    {
      "label": "Build Agent Service",
      "type": "shell",
      "command": "go",
      "args": [
        "build",
        "-o",
        "bin/jtnt-agentd.exe",
        "./cmd/agentd"
      ],
      "group": "build",
      "problemMatcher": ["$go"]
    },
    {
      "label": "Build MSI",
      "type": "shell",
      "command": "powershell",
      "args": [
        "-ExecutionPolicy",
        "Bypass",
        "-File",
        "${workspaceFolder}\\packaging\\windows\\build.ps1",
        "-Version",
        "4.0.0"
      ],
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "presentation": {
        "reveal": "always",
        "panel": "shared"
      },
      "problemMatcher": []
    },
    {
      "label": "Run Tests",
      "type": "shell",
      "command": "go",
      "args": ["test", "./..."],
      "group": "test",
      "problemMatcher": ["$go"]
    }
  ]
}
```

Press `Ctrl+Shift+B` to build MSI with one keystroke!

## VS Code Settings

Create `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "files.eol": "\n",
  "files.insertFinalNewline": true,
  "files.trimTrailingWhitespace": true,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "[powershell]": {
    "editor.defaultFormatter": "ms-vscode.PowerShell"
  }
}
```

## Quick Start Workflow

1. **Connect to portal.tsho.us**
   ```bash
   xfreerdp /v:portal.tsho.us /u:admin /size:1920x1080
   ```

2. **On Windows portal, open PowerShell as Admin**
   ```powershell
   # Install tools (one-time)
   choco install -y git golang wix311 vscode
   
   # Clone project
   cd C:\
   git clone https://github.com/tshojoshua/ai-support-system-agent.git jtnt-agent
   cd jtnt-agent
   
   # Open in VS Code
   code .
   ```

3. **In VS Code on Windows**
   - Install extensions listed above
   - Let Go extension install tools (2-3 min)
   - Open Copilot Chat: `Ctrl+Shift+I`
   - Paste the context from above
   - Ask: "Help me build the MSI installer"

4. **Build with one command**
   - Press `Ctrl+Shift+B` (runs Build MSI task)
   - OR: Open terminal and run:
     ```powershell
     cd packaging\windows
     .\build.ps1 -Version "4.0.0"
     ```

5. **Find MSI**
   ```
   packaging\windows\output\JTNT-Agent-4.0.0-x64.msi
   ```

## Copilot Chat Commands

Once Copilot has context, use these prompts:

```
# Start build
"Build the MSI installer following build.ps1"

# If build fails
"The build failed with error: [paste error]. How do I fix this?"

# Check prerequisites
"Verify all prerequisites are installed for WiX MSI build"

# Troubleshoot
"Check if Go modules are up to date"
"Verify WiX toolset is in PATH"

# Test after build
"Help me test the MSI installation silently with enrollment token"

# Understand code
"Explain how the auto-enrollment works in Product.wxs"
"What does the postinst script do?"
```

## Alternative: Use GitHub Copilot CLI

Install Copilot CLI on Windows:

```powershell
# Install GitHub CLI
winget install GitHub.cli

# Authenticate
gh auth login

# Install Copilot extension
gh extension install github/gh-copilot

# Use natural language
gh copilot suggest "build go windows binary for cmd/agentd"
gh copilot explain "build.ps1"
```

## File Locations Reference

Critical files you'll work with:
- `packaging/windows/build.ps1` - Main build script
- `packaging/windows/Product.wxs` - WiX MSI definition
- `packaging/windows/BUILD_GUIDE.md` - Detailed instructions
- `cmd/jtnt-agent/main.go` - CLI tool
- `cmd/agentd/main.go` - Service daemon
- `go.mod` - Go dependencies

## Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| Build (MSI) | `Ctrl+Shift+B` |
| Open Terminal | `Ctrl+` ` |
| Copilot Chat | `Ctrl+Shift+I` |
| Command Palette | `Ctrl+Shift+P` |
| Quick Open File | `Ctrl+P` |
| Go to Definition | `F12` |
| Find References | `Shift+F12` |

## Expected Build Output

```
Validating prerequisites...
✓ Go version: 1.23.4
✓ WiX Toolset found
✓ PowerShell: 5.1.19041.5247

Building binaries...
✓ bin/jtnt-agent.exe (5.2 MB)
✓ bin/jtnt-agentd.exe (5.4 MB)

Building MSI...
✓ Agent-4.0.0.wixobj
✓ JTNT-Agent-4.0.0-x64.msi (11.8 MB)

Build complete!
Location: packaging\windows\output\JTNT-Agent-4.0.0-x64.msi
```

## Troubleshooting

### Go extension not working
```powershell
# Manually install Go tools
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Build.ps1 execution policy error
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### WiX not in PATH
```powershell
# Add WiX to PATH
$env:Path += ";C:\Program Files (x86)\WiX Toolset v3.11\bin"
```

### Copilot not authenticated
1. Click Copilot icon in status bar
2. Sign in with GitHub account
3. Authorize VS Code

## Timeline Checkpoint

According to EMERGENCY_DEPLOYMENT.md:
- **Now → 11:30 PM**: Binaries built
- **11:30 PM → 12:15 AM**: MSI packaged
- **12:15 AM → 1:00 AM**: Testing
- **1:00 AM**: GO/NO-GO decision

Use Copilot to accelerate each phase!
