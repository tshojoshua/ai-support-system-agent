# JTNT Agent - Quick Start Guide

## What You Just Downloaded

Complete Phase 1 RMM agent source code (~2,000 lines) ready to build and deploy.

## File Structure

```
jtnt-agent/
├── cmd/                    # Executables
│   ├── agentd/            # Daemon (runs as service)
│   └── jtnt-agent/        # CLI tool (for enrollment, status)
├── internal/              # Core logic
│   ├── agent/            # Main orchestrator + heartbeat
│   ├── config/           # Configuration + OS-specific paths
│   ├── enroll/           # Enrollment flow
│   ├── store/            # Secure file storage
│   ├── sysinfo/          # System metrics collector
│   └── transport/        # HTTP client + retry logic
├── pkg/api/              # API types (shared with hub)
├── go.mod                # Dependencies
├── Makefile              # Build commands
└── README.md             # Full documentation
```

## Step-by-Step Setup

### 1. Prerequisites

Install Go 1.23 or later:
- **Windows**: https://go.dev/dl/
- **macOS**: `brew install go`
- **Linux**: `sudo apt install golang` or download from https://go.dev/dl/

### 2. Build

```bash
cd jtnt-agent

# Download dependencies
go mod download

# Build for your platform
make build

# Or build for all platforms
make build-all
```

**Output**: Binaries in `bin/` directory

### 3. Before You Can Test

You need an **enrollment token** from your hub. Generate one:

```bash
# On your hub server
cd /srv/posix-ai-hub/api
npx tsx scripts/generate-enrollment-token.ts \
  "<tenant-id>" \
  "<your-user-id>" \
  365 \
  10 \
  "<site-id-optional>" \
  "Test Agents"
```

**Note**: If this script doesn't exist yet, you need to implement Phase 5A first (hub enrollment endpoint).

### 4. Enroll

```bash
# Linux/macOS
./bin/jtnt-agent enroll --token <TOKEN> --hub https://hub.jtnt.us

# Windows
.\bin\jtnt-agent.exe enroll --token <TOKEN> --hub https://hub.jtnt.us
```

**What this does**:
1. Collects system info (hostname, CPU, memory, disk)
2. Sends enrollment request to hub
3. Receives agent_id and JWT token
4. Saves config to OS-specific location

### 5. Run

```bash
# Linux/macOS
sudo ./bin/jtnt-agentd

# Windows (as Administrator)
.\bin\jtnt-agentd.exe
```

**What it does**:
- Sends heartbeat every 60 seconds (default)
- Reports system metrics (CPU, memory, disk usage)
- Automatically retries on network failures

### 6. Check Status

```bash
./bin/jtnt-agent status
```

**Output**:
```
Status: Enrolled
Agent ID: abc-123-def-456
Hub URL: https://hub.jtnt.us
Tenant ID: xyz-789
Enrolled At: 2025-01-15 10:30:00
Config Path: /var/lib/jtnt-agent/config.json
```

## Testing

### Manual Test (Without Hub)

You can test the build without a hub:

```bash
# This will fail to enroll (no hub), but proves the binary works
./bin/jtnt-agent version
# Output: JTNT Agent v1.0.0

./bin/jtnt-agent status
# Output: Status: Not enrolled
```

### Full Integration Test

1. Generate enrollment token on hub
2. Enroll agent (step 4 above)
3. Run agent (step 5 above)
4. Check hub dashboard - you should see:
   - New agent registered
   - Heartbeats coming in every 60 seconds
   - System metrics updating

## Troubleshooting

### "Agent not enrolled"
Run the enrollment command first.

### "Connection refused"
- Check hub URL is correct
- Verify hub is running: `curl https://hub.jtnt.us/api/v1/health`
- Check firewall allows outbound HTTPS

### "Invalid token"
- Token may be expired
- Generate a new token
- Check you copied the entire token (no spaces)

### Logs

The agent logs to stdout. To capture:

```bash
# Linux/macOS
./bin/jtnt-agentd > agent.log 2>&1

# Or with systemd
journalctl -u jtnt-agentd -f
```

## What's Next

### Option 1: Test Phase 1 First
- Get enrollment working
- Verify heartbeats
- Then move to Phase 2 (job execution)

### Option 2: Add Hub Endpoint
You need to add the enrollment endpoint to your hub:

**File**: `/srv/posix-ai-hub/api/src/modules/agents/enrollment.routes.ts`

See the Phase 5A prompt I provided earlier for the complete code.

### Option 3: Continue to Phase 2
Once Phase 1 works, add:
- Job polling (`GET /api/v1/agents/jobs/next`)
- Job execution (run commands, scripts)
- Job results reporting

## Production Deployment

For production, you'll want:
1. **Phase 4**: Installers (MSI for Windows, PKG for macOS, DEB for Linux)
2. Service integration (systemd, launchd, Windows Service)
3. Auto-start on boot
4. Automatic updates

But test Phase 1 first!

## Support

Questions? Issues? Contact R3D at JTNT Communications.

---

**You're ready to build and test!** Start with `make build` and go from there.
