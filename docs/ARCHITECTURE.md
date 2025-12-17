# JTNT Agent Architecture - Phase 1

## Overview

The JTNT Agent is a secure, cross-platform Remote Monitoring and Management (RMM) agent designed to communicate with a central hub using mutual TLS (mTLS). Phase 1 focuses on establishing secure enrollment, persistent mTLS connections, and basic heartbeat monitoring.

## System Architecture

### High-Level Architecture

```
┌──────────────────────────────────────────────────┐
│                  JTNT Agent                      │
│                                                  │
│  ┌────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │    CLI     │  │   Daemon    │  │  Logger  │ │
│  │  jtnt-agent│  │ jtnt-agentd │  │   (JSON) │ │
│  └────────────┘  └─────────────┘  └──────────┘ │
│         │               │                │       │
│         └───────────────┴────────────────┘       │
│                         │                        │
│  ┌──────────────────────┴─────────────────────┐ │
│  │         Agent Core Orchestrator            │ │
│  │  - Lifecycle Management                    │ │
│  │  - Configuration                           │ │
│  │  - Heartbeat Coordination                  │ │
│  └────────────────────────────────────────────┘ │
│         │              │              │          │
│  ┌──────┴──────┐ ┌────┴─────┐ ┌─────┴──────┐  │
│  │  Transport  │ │ SysInfo  │ │   Store    │  │
│  │   (mTLS)    │ │Collector │ │  (Secure)  │  │
│  └─────────────┘ └──────────┘ └────────────┘  │
└──────────────────────────────────────────────────┘
                      ▲
                      │ mTLS (Client Cert)
                      │
                      ▼
         ┌────────────────────────┐
         │      JTNT Hub          │
         │  - Enrollment API      │
         │  - Heartbeat API       │
         │  - Job Distribution    │
         └────────────────────────┘
```

## Component Details

### 1. CLI Tool (`cmd/jtnt-agent`)

**Purpose**: User-facing command-line interface for agent management.

**Commands**:
- `enroll`: Initiate enrollment with hub using one-time token
- `status`: Display current agent state and configuration
- `version`: Show agent version
- `test-connection`: Verify mTLS connectivity to hub

**Flow**: CLI → Config/Enrollment/Transport → Result Display

### 2. Daemon (`cmd/agentd`)

**Purpose**: Long-running background process that maintains hub connection.

**Responsibilities**:
- Load configuration on startup
- Initialize agent core
- Start heartbeat loop
- Handle OS signals for graceful shutdown

**Lifecycle**:
```
Start → Load Config → Init Agent → Start Loops → Wait for Signal → Shutdown
```

### 3. Agent Core (`internal/agent`)

**Purpose**: Main orchestrator coordinating all agent activities.

**Components**:

#### agent.go
- Agent struct with dependencies (config, client, store, sysinfo)
- Start/Stop lifecycle management
- Context-based cancellation
- WaitGroup for goroutine coordination

#### heartbeat.go
- Periodic heartbeat transmission
- System info collection
- Dynamic interval adjustment based on hub response
- Error handling and retry

#### logger.go
- Structured JSON logging
- Log levels: debug, info, warn, error, fatal
- Context fields (agent_id, component, timestamp)

#### lifecycle.go
- Configuration reload
- Status reporting

### 4. Configuration (`internal/config`)

**Purpose**: Manage agent configuration and OS-specific paths.

**Files**:
- `config.go`: Config struct, load/save, validation
- `paths_linux.go`: Linux-specific paths
- `paths_darwin.go`: macOS-specific paths
- `paths_windows.go`: Windows-specific paths

**Storage Locations**:

| OS      | Config Path                                      | Certs Path                              |
|---------|--------------------------------------------------|----------------------------------------|
| Linux   | `/etc/jtnt-agent/config.json`                   | `/var/lib/jtnt-agent/certs/`          |
| macOS   | `/Library/Application Support/JTNT/Agent/config.json` | `/Library/Application Support/JTNT/Agent/certs/` |
| Windows | `C:\ProgramData\JTNT\Agent\config.json`        | `C:\ProgramData\JTNT\Agent\certs\`    |

### 5. Secure Storage (`internal/store`)

**Purpose**: Platform-specific secure file storage with appropriate permissions.

**Interface**:
```go
type Store interface {
    Save(key string, data []byte) error
    Load(key string) ([]byte, error)
    Exists(key string) bool
    Delete(key string) error
    SetPermissions(path string) error
}
```

**Platform Implementations**:
- **Linux/macOS**: File permissions 0600, ownership validation
- **Windows**: ACL-based permissions, owner-only access

### 6. Enrollment (`internal/enroll`)

**Purpose**: Handle initial agent enrollment with hub.

**Process Flow**:

```
1. Generate Ed25519 Keypair
   ↓
2. Collect System Info (hostname, OS, arch)
   ↓
3. POST /api/v1/agent/enroll
   - Token
   - Public Key
   - System Info
   ↓
4. Receive Response
   - Agent ID
   - Client Certificate
   - Client Key
   - CA Bundle
   - Policy
   ↓
5. Validate Certificate Chain
   ↓
6. Save Certificates (secure storage)
   ↓
7. Save Configuration
   ↓
8. Enrollment Complete
```

**Security**:
- Ed25519 keypair never leaves agent
- Certificate chain validation before acceptance
- Secure storage with restrictive permissions

### 7. Transport (`internal/transport`)

**Purpose**: mTLS HTTP client with retry logic.

#### client.go
- mTLS configuration with client certificates
- Connection pooling
- Request/response handling
- Status code classification

#### retry.go
- Exponential backoff: 30s → 15min
- Jitter: ±20%
- Retryable vs non-retryable error classification

**Error Classification**:

| Category       | Examples                          | Action |
|----------------|-----------------------------------|--------|
| Retryable      | Connection refused, timeout, DNS  | Retry  |
| Non-retryable  | Certificate errors, 4xx (except 429) | Fail |
| Rate limited   | 429 Too Many Requests             | Retry  |

### 8. System Info (`internal/sysinfo`)

**Purpose**: Collect system metrics for heartbeat.

**Metrics Collected**:
- Hostname
- OS and version
- Architecture
- Uptime
- CPU count and usage
- Memory total/used
- Disk total/used
- IP addresses (non-loopback IPv4)

**Library**: Uses `gopsutil/v3` for cross-platform metrics.

## Data Flow

### Enrollment Flow

```
User
  │
  ├─ jtnt-agent enroll --token TOKEN --hub URL
  │
  ▼
Enroller
  │
  ├─ Generate Ed25519 Keypair
  ├─ Collect hostname, OS, arch
  │
  ▼
POST /api/v1/agent/enroll
  │
  ▼
Hub Response
  │
  ├─ agent_id
  ├─ client_cert_pem
  ├─ client_key_pem
  ├─ ca_bundle_pem
  ├─ policy
  │
  ▼
Validate Certificates
  │
  ├─ Parse client cert
  ├─ Parse CA bundle
  ├─ Verify chain
  │
  ▼
Save to Secure Storage
  │
  ├─ client.crt (0600)
  ├─ client.key (0600)
  ├─ ca-bundle.crt (0600)
  ├─ config.json (0600)
  │
  ▼
Enrollment Complete
```

### Heartbeat Flow

```
Ticker (every N seconds)
  │
  ▼
Collect System Info
  │
  ├─ CPU, memory, disk metrics
  ├─ Network interfaces
  │
  ▼
Create HeartbeatRequest
  │
  ├─ agent_id
  ├─ timestamp
  ├─ sysinfo
  │
  ▼
POST /api/v1/agent/heartbeat (mTLS)
  │
  ├─ Retry on failure
  ├─ Exponential backoff
  │
  ▼
Hub Response
  │
  ├─ ok: true
  ├─ next_heartbeat_sec
  │
  ▼
Update Interval (if changed)
  │
  ▼
Log Success/Failure
```

## Security Considerations

### Enrollment Security

1. **One-Time Token**: Enrollment token is single-use and time-limited
2. **Ed25519**: Modern, secure elliptic curve cryptography
3. **Certificate Validation**: Full chain validation before acceptance
4. **Secure Storage**: Certificates stored with restrictive OS permissions

### Runtime Security

1. **mTLS**: All communication after enrollment requires valid client certificate
2. **Certificate Pinning**: CA bundle from enrollment used for validation
3. **No Plaintext Secrets**: All sensitive data encrypted or access-controlled
4. **Least Privilege**: Agent runs with minimum required permissions

### Transport Security

1. **TLS 1.2+**: Minimum TLS version enforced
2. **Certificate Expiration**: Future phases will handle rotation
3. **Connection Pooling**: Reuse of validated connections
4. **Timeout Enforcement**: All requests have context timeouts

## Error Handling

### Strategy

1. **Graceful Degradation**: Agent continues running even if heartbeat fails
2. **Exponential Backoff**: Prevent hub overload during outages
3. **Structured Logging**: All errors logged with context
4. **Non-Fatal Errors**: Most errors don't crash agent

### Retry Policy

```
Initial:  30 seconds
Maximum:  15 minutes
Jitter:   ±20%
Formula:  backoff = min(30s * 2^attempt, 15min) ± jitter
```

## Concurrency Model

### Goroutines

1. **Main Goroutine**: Manages lifecycle, handles signals
2. **Heartbeat Loop**: Periodic heartbeat transmission
3. **Future**: Job polling (Phase 2+)

### Synchronization

- **Context**: Cancellation propagation for shutdown
- **WaitGroup**: Ensures all goroutines complete before exit
- **Mutex**: (Not needed in Phase 1, future phases for shared state)

## Testing Strategy

### Unit Tests

- Configuration loading/saving
- Certificate validation
- Retry logic with mock errors
- System info collection
- Backoff calculation

### Integration Tests (Future)

- End-to-end enrollment
- Heartbeat with test hub
- Certificate rotation

### Platform Tests

- Build verification on Windows, macOS, Linux
- Path resolution on each OS
- Permission setting verification

## Performance Characteristics

### Resource Usage

- **Memory**: ~10-20 MB baseline
- **CPU**: <1% during idle, <5% during heartbeat
- **Network**: Periodic heartbeats (default 60s), minimal bandwidth
- **Disk**: Config + certs (~10 KB)

### Scalability

- **Heartbeat**: Designed for 10,000+ agents per hub
- **Backoff**: Prevents thundering herd during outages
- **Connection Pooling**: Reduces connection overhead

## Future Phases

### Phase 2: Job Execution
- Job polling
- Command execution
- Output capture and reporting

### Phase 3: Policy Enforcement
- Capability-based restrictions
- Audit logging
- Compliance checks

### Phase 4: Service Management
- Native OS service installation
- Auto-update mechanism
- Certificate rotation

### Phase 5: Advanced Features
- Plugin system
- Custom metrics
- Local caching

## Monitoring and Observability

### Logging

- **Format**: Structured JSON
- **Destination**: stdout (systemd journal, launchd log)
- **Levels**: debug, info, warn, error, fatal
- **Rotation**: Handled by OS service manager

### Metrics (Future)

- Heartbeat success rate
- Response time
- Certificate expiration countdown
- System resource usage

## Compliance and Audit

### Data Collection

- **PII**: Hostname and IP addresses only
- **System Metrics**: Aggregated, non-identifying
- **Logs**: Local only, not transmitted

### Certificate Storage

- Restricted OS permissions
- Owner-only read/write access
- Validated before use

---

**Document Version**: 1.0.0  
**Phase**: 1 (Core Foundation)  
**Last Updated**: December 16, 2025
