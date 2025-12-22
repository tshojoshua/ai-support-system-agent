# JTNT RMM Agent

Cross-platform remote management agent for JTNT Hub (hub.jtnt.us).

## Features

- ✅ Secure enrollment with JWT tokens
- ✅ Periodic heartbeat with system metrics (CPU, memory, disk)
- ✅ Cross-platform support (Windows, macOS, Linux)
- ✅ OS-specific secure storage
- ✅ Automatic retry with exponential backoff
- ✅ Multi-tenancy support

## Quick Start

### 1. Build

```bash
# Install dependencies
go mod download

# Build for current platform
make build

# Build for all platforms
make build-all
```

Binaries will be in the `bin/` directory.

### 2. Enroll

Get an enrollment token from the hub administrator, then:

**Linux/macOS:**
```bash
./bin/jtnt-agent enroll --token <YOUR_TOKEN> --hub https://hub.jtnt.us
```

**Windows:**
```cmd
.\bin\jtnt-agent.exe enroll --token <YOUR_TOKEN> --hub https://hub.jtnt.us
```

### 3. Run

**Linux/macOS:**
```bash
sudo ./bin/jtnt-agentd
```

**Windows (as Administrator):**
```cmd
.\bin\jtnt-agentd.exe
```

### 4. Check Status

```bash
./bin/jtnt-agent status
```

## Installation Paths

Configuration is stored at:
- **Windows**: `C:\ProgramData\JTNT\Agent\config.json`
- **macOS**: `/Library/Application Support/JTNT/Agent/config.json`
- **Linux**: `/var/lib/jtnt-agent/config.json`

## System Requirements

- **Windows**: Windows 10 or later, Windows Server 2019 or later
- **macOS**: macOS 13 (Ventura) or later
- **Linux**: Ubuntu 20.04+, Debian 11+, or equivalent

## Commands

```bash
# Enroll the agent
jtnt-agent enroll --token <TOKEN> --hub <URL>

# Check agent status
jtnt-agent status

# Show version
jtnt-agent version
```

## Development

```bash
# Run tests
make test

# Clean build artifacts
make clean

# Development run (builds and starts daemon)
make dev
```

## Architecture

```
jtnt-agent/
├── cmd/
│   ├── agentd/          # Main daemon
│   └── jtnt-agent/      # CLI tool
├── internal/
│   ├── agent/           # Core agent logic
│   ├── config/          # Configuration management
│   ├── enroll/          # Enrollment logic
│   ├── store/           # Secure storage
│   ├── sysinfo/         # System information collector
│   └── transport/       # HTTP client with retry
└── pkg/
    └── api/             # API types
```

## What's Collected

The agent periodically sends system information to the hub:
- Hostname
- OS and version
- CPU count and usage
- Memory total and usage
- Disk total and usage
- IP addresses
- Uptime

## Security

- Configuration stored with restricted permissions (0600)
- JWT-based authentication
- HTTPS-only communication
- No inbound ports required (outbound only)

## Roadmap

- **Phase 1**: ✅ Core agent (enrollment, heartbeat) - **COMPLETE**
- **Phase 2**: Job execution engine (remote commands, scripts)
- **Phase 3**: Metrics, health checks, self-update
- **Phase 4**: Installers (MSI, PKG, DEB)

## Support

For issues or questions, contact JTNT Communications support.

## License

Proprietary - JTNT Communications
