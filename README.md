# JTNT RMM Agent - Phase 1

A secure, cross-platform Remote Monitoring and Management (RMM) agent written in Go. This is Phase 1 implementation focusing on core enrollment, mTLS transport, and basic agent lifecycle.

## Features

- **Secure Enrollment**: One-time token-based enrollment with Ed25519 keypair generation
- **mTLS Transport**: All communication secured with mutual TLS authentication
- **Cross-Platform**: Supports Windows, macOS (Intel & Apple Silicon), and Linux
- **Heartbeat Monitoring**: Periodic system health reporting
- **Structured Logging**: JSON-formatted logs for easy parsing
- **Retry Logic**: Exponential backoff with jitter for network resilience

## Architecture

```
┌─────────────┐         mTLS          ┌──────────┐
│ JTNT Agent  │◄─────────────────────►│   Hub    │
│             │    Heartbeat/Jobs      │          │
└─────────────┘                        └──────────┘
      │
      ├─ Config & Certs (secure storage)
      ├─ System Info Collection
      └─ Job Execution (future phases)
```

## Building

### Prerequisites

- Go 1.23 or later
- Make (optional but recommended)

### Quick Build

```bash
# Install dependencies
make deps

# Build for current platform
make build

# Build for all platforms
make build-all
```

Binaries will be in the `bin/` directory.

### Manual Build

```bash
# Daemon
go build -o bin/jtnt-agentd ./cmd/agentd

# CLI
go build -o bin/jtnt-agent ./cmd/jtnt-agent
```

## Installation

### Linux/macOS

```bash
# Build and install
make build
sudo make install

# This installs to /usr/local/bin/
```

### Windows

1. Build for Windows: `make build-windows`
2. Copy `bin/windows/*.exe` to desired location (e.g., `C:\Program Files\JTNT\Agent\`)
3. Add to PATH or run from installation directory

## Usage

### 1. Enrollment

First, enroll the agent with your JTNT hub:

```bash
# Linux/macOS
sudo jtnt-agent enroll --token YOUR_ENROLLMENT_TOKEN --hub https://hub.jtnt.us

# Windows (as Administrator)
jtnt-agent.exe enroll --token YOUR_ENROLLMENT_TOKEN --hub https://hub.jtnt.us
```

This will:
- Generate an Ed25519 keypair
- Exchange the token for agent credentials
- Receive and validate mTLS certificates
- Save configuration and certificates securely

### 2. Check Status

```bash
jtnt-agent status
```

Output:
```
Agent Status:
  Agent ID:         550e8400-e29b-41d4-a716-446655440000
  Hub URL:          https://hub.jtnt.us
  Heartbeat:        60s
  Poll Interval:    300s
  Policy Version:   1
  Config File:      /etc/jtnt-agent/config.json
  Cert Path:        /var/lib/jtnt-agent/certs/client.crt
```

### 3. Test Connection

```bash
jtnt-agent test-connection
```

### 4. Run Agent

#### Manual Mode

```bash
# Linux/macOS
sudo jtnt-agentd

# Windows (as Administrator)
jtnt-agentd.exe
```

#### As a Service

**Linux (systemd)**

Create `/etc/systemd/system/jtnt-agent.service`:

```ini
[Unit]
Description=JTNT RMM Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/jtnt-agentd
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable jtnt-agent
sudo systemctl start jtnt-agent
sudo systemctl status jtnt-agent
```

**macOS (launchd)**

Create `/Library/LaunchDaemons/us.jtnt.agent.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>us.jtnt.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/jtnt-agentd</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/jtnt-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/jtnt-agent.error.log</string>
</dict>
</plist>
```

Then:
```bash
sudo launchctl load /Library/LaunchDaemons/us.jtnt.agent.plist
sudo launchctl start us.jtnt.agent
```

**Windows (Service)**

Service installation will be added in Phase 4.

## File Locations

### Linux

- Binary: `/usr/local/bin/jtnt-agentd`
- Config: `/etc/jtnt-agent/config.json`
- State: `/var/lib/jtnt-agent/`
- Certs: `/var/lib/jtnt-agent/certs/`

### macOS

- Binary: `/usr/local/bin/jtnt-agentd`
- Config: `/Library/Application Support/JTNT/Agent/config.json`
- State: `/Library/Application Support/JTNT/Agent/`
- Certs: `/Library/Application Support/JTNT/Agent/certs/`

### Windows

- Binary: `C:\Program Files\JTNT\Agent\jtnt-agentd.exe`
- Config: `C:\ProgramData\JTNT\Agent\config.json`
- State: `C:\ProgramData\JTNT\Agent\`
- Certs: `C:\ProgramData\JTNT\Agent\certs\`

## Configuration

Configuration is stored in `config.json`:

```json
{
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "hub_url": "https://hub.jtnt.us",
  "poll_interval_sec": 300,
  "heartbeat_sec": 60,
  "cert_path": "/var/lib/jtnt-agent/certs/client.crt",
  "key_path": "/var/lib/jtnt-agent/certs/client.key",
  "ca_bundle_path": "/var/lib/jtnt-agent/certs/ca-bundle.crt",
  "policy_version": 1
}
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run static analysis
make vet

# Run linter (requires golangci-lint)
make lint
```

### Project Structure

```
jtnt-agent/
├── cmd/
│   ├── agentd/          # Daemon entry point
│   └── jtnt-agent/      # CLI tool
├── internal/
│   ├── agent/           # Main agent orchestrator
│   ├── config/          # Configuration management
│   ├── enroll/          # Enrollment logic
│   ├── store/           # Secure storage
│   ├── sysinfo/         # System information
│   └── transport/       # mTLS HTTP client
├── pkg/
│   └── api/             # Shared API types
├── Makefile
└── README.md
```

## API Endpoints

### Enrollment

```
POST /api/v1/agent/enroll
Content-Type: application/json

{
  "token": "enrollment-token",
  "hostname": "client-machine",
  "os": "linux",
  "arch": "amd64",
  "version": "1.0.0",
  "public_key": "base64-ed25519-pubkey"
}
```

### Heartbeat

```
POST /api/v1/agent/heartbeat
Content-Type: application/json
mTLS Required

{
  "agent_id": "uuid",
  "timestamp": "2025-01-15T10:30:00Z",
  "sysinfo": { ... }
}
```

## Security

- **mTLS**: All post-enrollment communication uses mutual TLS
- **Certificate Validation**: Full chain validation with CA bundle
- **Secure Storage**: OS-specific permissions (0600 on Unix, ACLs on Windows)
- **Ed25519**: Modern elliptic curve cryptography for keypairs
- **No Plaintext Secrets**: All sensitive data encrypted or secured

## Logging

Structured JSON logs to stdout:

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "level": "info",
  "component": "heartbeat",
  "agent_id": "uuid",
  "message": "heartbeat sent successfully",
  "fields": {
    "duration_ms": 234
  }
}
```

Levels: `debug`, `info`, `warn`, `error`, `fatal`

## Troubleshooting

### Agent won't enroll

1. Check network connectivity to hub
2. Verify enrollment token is valid and not expired
3. Check firewall rules allow HTTPS outbound
4. Review enrollment logs for specific errors

### Agent won't start

1. Verify enrollment completed: `jtnt-agent status`
2. Check certificates exist and are readable
3. Verify config file exists: `/etc/jtnt-agent/config.json` (Linux)
4. Check file permissions on certs directory

### Connection issues

1. Test connection: `jtnt-agent test-connection`
2. Verify hub URL is reachable
3. Check certificate expiration
4. Review agent logs for network errors

## Phase 1 Limitations

This is Phase 1 implementation. The following features are **not yet implemented**:

- Job execution and command processing
- Policy enforcement beyond basic capabilities
- Self-update mechanism
- Advanced system metrics
- Windows service installation

These will be added in subsequent phases.

## License

Copyright © 2025 JTNT. All rights reserved.

## Support

For issues and questions:
- GitHub Issues: https://github.com/tshojoshua/jtnt-agent/issues
- Documentation: https://docs.jtnt.us

## Version

Current version: **1.0.0** (Phase 1)
