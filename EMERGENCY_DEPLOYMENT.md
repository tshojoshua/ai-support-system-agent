# 🚨 EMERGENCY: Windows MSI Build - Deploy TONIGHT

**Priority:** 🔴 **CRITICAL - CLIENT WAITING**  
**Timeline:** 2.5 hours MAX to complete MSI + basic testing  
**Current Time:** 10:30 PM PST  
**Deadline:** 1:00 AM (MSI ready for deployment)  

---

## SITUATION

Hub backend is READY (all endpoints operational). Windows MSI is the ONLY blocker. Client deployment scheduled for 1:30 AM tonight. NO TIME FOR MISTAKES.

**Your Mission:** Build and basic-test Windows MSI in 2.5 hours.

---

## COMPRESSED TIMELINE

### Hour 1: Setup + Build (10:30 PM - 11:30 PM)

**10:30 - 10:45 (15 min): Windows Machine Setup**
```powershell
# OPTION 1: Use your Windows laptop/desktop
# Check you have admin access
net session >nul 2>&1 && echo Admin: YES || echo Admin: NO

# Verify OS: Windows 10/11 or Server 2019/2022
Get-ComputerInfo | Select WindowsVersion, OsArchitecture

# OPTION 2: Spin up Azure/AWS Windows VM (if no local machine)
# Use t3.medium (AWS) or Standard_D2s_v3 (Azure)
# RDP in immediately
```

**10:45 - 11:00 (15 min): Install Build Tools**
```powershell
# Run PowerShell as Administrator
Set-ExecutionPolicy Bypass -Scope Process -Force

[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install required tools (10 minutes)
choco install -y git golang wix311

# Verify (should all work)
git --version
go version
candle.exe -help
light.exe -help
```

**11:00 - 11:10 (10 min): Get Agent Code**
```powershell
# Clone repository
cd C:\
git clone https://github.com/tshojoshua/ai-support-system-agent.git jtnt-agent
cd jtnt-agent
```

**11:10 - 11:30 (20 min): Build Agent Binaries**
```powershell
cd C:\jtnt-agent

# Set version
$VERSION = "4.0.0"

# Create bin directory
New-Item -ItemType Directory -Force -Path bin

# Build CLI
Write-Host "Building jtnt-agent.exe..." -ForegroundColor Cyan
go build -o bin\jtnt-agent.exe -ldflags "-s -w -X main.Version=$VERSION" .\cmd\jtnt-agent

# Build daemon
Write-Host "Building jtnt-agentd.exe..." -ForegroundColor Cyan
go build -o bin\jtnt-agentd.exe -ldflags "-s -w -X main.Version=$VERSION" .\cmd\agentd

# Verify
ls bin\*.exe | ForEach-Object {
    "{0}: {1:N2} MB" -f $_.Name, ($_.Length / 1MB)
}
# Should see 2 EXE files, 10-15 MB each

# If build fails with missing packages:
go mod download
go mod tidy
# Then retry build
```

**⚠️ CHECKPOINT 1 (11:30 PM): GO/NO-GO**
- ✅ Binaries exist: `bin\jtnt-agent.exe` + `bin\jtnt-agentd.exe`
- ✅ Each file 10-15 MB
- ✅ No build errors
- **IF FAIL:** Escalate immediately, consider aborting to tomorrow

---

### Hour 2: Package MSI (11:30 PM - 12:30 AM)

**11:30 - 12:15 (45 min): Build MSI**
```powershell
cd C:\jtnt-agent\packaging\windows

# Create output directories
New-Item -ItemType Directory -Force -Path output
New-Item -ItemType Directory -Force -Path obj

# Compile WiX source (Product.wxs already exists)
Write-Host "Compiling WiX source..." -ForegroundColor Cyan
candle.exe Product.wxs -dAgentPath=..\..\bin -arch x64 -out obj\

# Check for errors
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: candle.exe failed!" -ForegroundColor Red
    Write-Host "Check Product.wxs for syntax errors" -ForegroundColor Yellow
    exit 1
}

# Link to MSI
Write-Host "Linking MSI..." -ForegroundColor Cyan
light.exe obj\Product.wixobj -out output\JTNT-Agent-4.0.0.msi -ext WixUIExtension -cultures:en-us -loc en-us.wxl

# Verify MSI created
if (Test-Path output\JTNT-Agent-4.0.0.msi) {
    $size = (Get-Item output\JTNT-Agent-4.0.0.msi).Length / 1MB
    Write-Host "SUCCESS: MSI created: {0:N2} MB" -f $size -ForegroundColor Green
} else {
    Write-Host "ERROR: MSI not created!" -ForegroundColor Red
    exit 1
}
```

**⚠️ CHECKPOINT 2 (12:15 AM): GO/NO-GO**
- ✅ MSI file exists: `output\JTNT-Agent-4.0.0.msi`
- ✅ File size 15-25 MB
- ✅ No build errors
- **IF FAIL:** CRITICAL BLOCKER - Escalate immediately

**12:15 - 12:30 (15 min): Package for Deployment**
```powershell
# Create deployment folder
New-Item -ItemType Directory -Force -Path C:\deployment

# Copy MSI
Copy-Item output\JTNT-Agent-4.0.0.msi C:\deployment\

# Copy helper scripts
Copy-Item install.bat C:\deployment\
Copy-Item install-silent.bat C:\deployment\
Copy-Item uninstall.bat C:\deployment\

# Calculate checksum
Get-FileHash C:\deployment\JTNT-Agent-4.0.0.msi -Algorithm SHA256 | 
    Select-Object Algorithm, Hash | 
    Format-List | 
    Out-File C:\deployment\checksum.txt

# Create deployment ZIP
Compress-Archive -Path C:\deployment\* -DestinationPath C:\deployment\JTNT-Agent-Windows-4.0.0.zip -Force

# Display results
Write-Host "`nDeployment Package Ready:" -ForegroundColor Green
ls C:\deployment\JTNT-Agent-* | ForEach-Object {
    "{0}: {1:N2} MB" -f $_.Name, ($_.Length / 1MB)
}
```

---

### Hour 3: RAPID TESTING (12:30 AM - 1:00 AM)

**12:30 - 12:50 (20 min): Quick Install Test**

**BEFORE TESTING: Get test token from Hub admin**

```powershell
# Set test token (GET THIS FROM HUB ADMIN)
$TEST_TOKEN = "etok_test_emergency_dec21"  # REPLACE WITH ACTUAL TOKEN

# Silent install
Write-Host "Installing agent..." -ForegroundColor Cyan
msiexec /i C:\deployment\JTNT-Agent-4.0.0.msi /qn /l*v C:\deployment\install-test.log `
    ENROLLMENT_TOKEN=$TEST_TOKEN `
    HUB_URL="https://hub.jtnt.us"

# Wait for installation
Write-Host "Waiting 30 seconds for installation..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# CRITICAL CHECK 1: Service running?
Write-Host "`n=== CHECK 1: Service Status ===" -ForegroundColor Cyan
$service = Get-Service JTNTAgent -ErrorAction SilentlyContinue
if ($service -and $service.Status -eq 'Running') {
    Write-Host "✅ Service is Running" -ForegroundColor Green
} else {
    Write-Host "❌ Service NOT running!" -ForegroundColor Red
    Write-Host "Status: $($service.Status)" -ForegroundColor Yellow
}

# CRITICAL CHECK 2: Agent enrolled?
Write-Host "`n=== CHECK 2: Agent Status ===" -ForegroundColor Cyan
if (Test-Path "C:\Program Files\JTNT\Agent\jtnt-agent.exe") {
    $status = & "C:\Program Files\JTNT\Agent\jtnt-agent.exe" status
    Write-Host $status
    
    if ($status -match "enrolled.*true" -or $status -match "Agent ID") {
        Write-Host "✅ Agent enrolled successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Agent NOT enrolled!" -ForegroundColor Red
    }
} else {
    Write-Host "❌ Agent binary not found!" -ForegroundColor Red
}

# CRITICAL CHECK 3: Logs clean?
Write-Host "`n=== CHECK 3: Agent Logs ===" -ForegroundColor Cyan
$logPath = "C:\ProgramData\JTNT\Agent\logs\agent.log"
if (Test-Path $logPath) {
    $logs = Get-Content $logPath -Tail 30
    
    # Check for errors
    $errors = $logs | Select-String -Pattern "error|fail|404" -CaseSensitive:$false
    if ($errors) {
        Write-Host "⚠️  Errors found in logs:" -ForegroundColor Yellow
        $errors | ForEach-Object { Write-Host $_ -ForegroundColor Red }
    } else {
        Write-Host "✅ No errors in logs" -ForegroundColor Green
    }
    
    # Show recent logs
    Write-Host "`nRecent log entries:" -ForegroundColor Cyan
    $logs[-10..-1] | ForEach-Object { Write-Host $_ -ForegroundColor Gray }
} else {
    Write-Host "❌ Log file not found!" -ForegroundColor Red
}

# CRITICAL CHECK 4: Install log errors?
Write-Host "`n=== CHECK 4: Installation Log ===" -ForegroundColor Cyan
if (Test-Path C:\deployment\install-test.log) {
    $installErrors = Select-String -Path C:\deployment\install-test.log -Pattern "error|failed" -CaseSensitive:$false
    if ($installErrors) {
        Write-Host "⚠️  Installation errors found:" -ForegroundColor Yellow
        $installErrors | Select-Object -First 5 | ForEach-Object { Write-Host $_.Line -ForegroundColor Red }
    } else {
        Write-Host "✅ No installation errors" -ForegroundColor Green
    }
}
```

**12:50 - 1:00 (10 min): Verify Hub Integration**

**Contact Hub Admin to verify:**
```bash
# Check agent appears in Hub dashboard
# https://hub.jtnt.us/agents

# Check heartbeat in Hub logs
docker compose logs hub-api | grep -i heartbeat | tail -20

# Check agent in database
docker compose exec hub-db psql -U postgres -d hub -c \
  "SELECT id, hostname, status, last_heartbeat FROM agents ORDER BY created_at DESC LIMIT 1;"
```

**⚠️ CHECKPOINT 3 (1:00 AM): FINAL GO/NO-GO DECISION**

**ALL MUST PASS:**
- ✅ MSI installs without errors
- ✅ Service starts and runs (Status = Running)
- ✅ Agent enrolls successfully (shows Agent ID)
- ✅ Heartbeat visible in Hub (check Hub logs/dashboard)
- ✅ No critical errors in agent logs
- ✅ Hub admin confirms agent visible

**IF ANY FAIL:**
```powershell
# Uninstall test agent
msiexec /x C:\deployment\JTNT-Agent-4.0.0.msi /qn

Write-Host "`n🔴 ABORT DEPLOYMENT - RESCHEDULE TO TOMORROW" -ForegroundColor Red
Write-Host "Document failure reason and notify all stakeholders" -ForegroundColor Yellow
```

**IF ALL PASS:**
```powershell
Write-Host "`n🟢 GO FOR CLIENT DEPLOYMENT" -ForegroundColor Green
Write-Host "Proceed to client deployment at 1:30 AM" -ForegroundColor Cyan
```

---

## CRITICAL SUCCESS CRITERIA

**ALL must pass to deploy to client:**

1. ✅ MSI file exists and is 15-25 MB
2. ✅ Installs on test Windows machine without errors
3. ✅ Service starts automatically
4. ✅ Agent enrolls with token
5. ✅ Agent shows in Hub dashboard
6. ✅ Heartbeat working (visible in Hub logs)
7. ✅ No errors in agent.log or install log

**If ANY fail → STOP → Reschedule to tomorrow**

---

## EMERGENCY TROUBLESHOOTING

### Service Won't Start
```powershell
# Check service details
sc.exe query JTNTAgent
sc.exe qc JTNTAgent

# Check Event Viewer
Get-EventLog -LogName Application -Source "JTNT Agent" -Newest 10

# Try manual start
net start JTNTAgent

# If fails, check binary exists
Test-Path "C:\Program Files\JTNT\Agent\jtnt-agentd.exe"
```

### Enrollment Fails (404)
```powershell
# Test Hub API endpoint
curl -X POST https://hub.jtnt.us/api/v1/agent/enroll -H "Content-Type: application/json" -d '{"token":"test"}'

# If 404: Hub API not deployed (ABORT)
# If 400/401: Endpoint exists, token issue (check token)
```

### Build Fails
```powershell
# Clear Go cache
go clean -modcache
go mod download
go mod tidy

# Retry build
go build -o bin\jtnt-agent.exe .\cmd\jtnt-agent
go build -o bin\jtnt-agentd.exe .\cmd\agentd
```

---

## COMMUNICATION PROTOCOL

**Status Updates (Send to coordination channel):**
- **11:00 PM**: "Build tools installed, binaries compiling"
- **11:30 PM**: "Binaries complete, packaging MSI"
- **12:00 AM**: "MSI build in progress"
- **12:30 AM**: "MSI complete, starting tests"
- **1:00 AM**: "Testing complete - GO/NO-GO: [DECISION]"

**Format:**
```
TIME: [HH:MM AM]
STATUS: [On Track / Issues / Blocked]
DETAILS: [Brief description]
NEXT: [Next milestone]
ETA: [If delayed]
```

---

## DELIVERABLES (BY 1:00 AM)

**Files:**
1. ✅ `JTNT-Agent-4.0.0.msi` (15-25 MB)
2. ✅ `install.bat` (installation helper)
3. ✅ `install-silent.bat` (silent deployment)
4. ✅ `uninstall.bat` (uninstall helper)
5. ✅ `checksum.txt` (SHA256 hash)
6. ✅ `JTNT-Agent-Windows-4.0.0.zip` (complete package)

**Test Results:**
```
Test Installation Report - Dec 21, 2025 1:00 AM

MSI Build: ✅ PASS / ❌ FAIL
Installation: ✅ PASS / ❌ FAIL
Service Start: ✅ PASS / ❌ FAIL
Enrollment: ✅ PASS / ❌ FAIL
Heartbeat: ✅ PASS / ❌ FAIL
Hub Visibility: ✅ PASS / ❌ FAIL

DECISION: 🟢 GO / 🔴 NO-GO

Notes:
[Any issues or concerns]
```

---

## ABORT CRITERIA

**STOP IMMEDIATELY IF:**
- Cannot get Windows machine with admin access in 30 minutes
- Build tools won't install after 20 minutes
- Agent binaries fail to compile after 30 minutes
- MSI build fails after 45 minutes
- Test installation fails any critical check
- Hub admin cannot verify agent in dashboard
- Team confidence < 95%

**NO SHAME IN ABORTING. Better to delay than fail at client.**

---

## FINAL REMINDERS

**This is EMERGENCY MODE:**
- ⏱️ Time is critical - work efficiently
- 📢 Communicate every 30 minutes
- ✅ All checkpoints must pass
- 🚫 Don't skip testing
- 💯 95% confidence required to proceed

**Client has LOW TOLERANCE:**
- Perfect installation or reschedule
- No partial deployments
- No "we'll fix it later"

**You have 2.5 hours. GO NOW.**

---

**Start Time:** 10:30 PM PST  
**Checkpoint 1:** 11:30 PM (binaries)  
**Checkpoint 2:** 12:15 AM (MSI)  
**Checkpoint 3:** 1:00 AM (GO/NO-GO)  
**Deployment:** 1:30 AM (if GO)
