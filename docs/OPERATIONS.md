# JTNT Agent Operations Guide

## Table of Contents
1. [Monitoring and Metrics](#monitoring-and-metrics)
2. [Health Checks](#health-checks)
3. [Certificate Management](#certificate-management)
4. [Updates](#updates)
5. [Troubleshooting](#troubleshooting)
6. [Audit Logs](#audit-logs)

## Monitoring and Metrics

### Prometheus Metrics

The agent exposes Prometheus metrics on `http://localhost:9090/metrics`.

**Important**: Metrics endpoint is localhost-only for security.

#### Available Metrics

**Counters:**
```
jtnt_agent_heartbeat_total{status="success|error"}
jtnt_agent_jobs_executed_total{type="exec|script|download|upload", status="success|error|timeout"}
jtnt_agent_enrollment_attempts_total{status="success|error"}
jtnt_agent_policy_violations_total{type="exec|script|file"}
jtnt_agent_update_attempts_total{status="success|error"}
jtnt_agent_cert_rotation_total{status="success|error"}
```

**Gauges:**
```
jtnt_agent_up{version="x.x.x"}
jtnt_agent_heartbeat_last_success_timestamp
jtnt_agent_job_execution_active
jtnt_agent_hub_connection_status{status="connected|disconnected"}
jtnt_agent_policy_expiration_timestamp
jtnt_agent_cert_expiration_timestamp
jtnt_agent_system_cpu_usage_percent
jtnt_agent_system_memory_used_bytes
jtnt_agent_system_disk_used_bytes
```

**Histograms:**
```
jtnt_agent_heartbeat_duration_seconds
jtnt_agent_job_execution_duration_seconds{type="exec|script|download|upload"}
jtnt_agent_artifact_upload_duration_seconds
jtnt_agent_api_request_duration_seconds{endpoint="/api/v1/..."}
```

### Scraping Metrics

**Prometheus configuration:**

```yaml
scrape_configs:
  - job_name: 'jtnt-agent'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 30s
```

### Key Metrics to Alert On

1. **Agent Down**: `jtnt_agent_up == 0`
2. **Hub Disconnected**: `jtnt_agent_hub_connection_status{status="disconnected"} == 1`
3. **Certificate Expiring**: `(jtnt_agent_cert_expiration_timestamp - time()) < 2592000` (30 days)
4. **Policy Expired**: `jtnt_agent_policy_expiration_timestamp < time()`
5. **High Disk Usage**: `jtnt_agent_system_disk_used_bytes > threshold`

## Health Checks

### Health Check Endpoint

The agent exposes a health check endpoint on `http://localhost:9091/health`.

**Healthy Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-12-16T10:30:00Z",
  "checks": {
    "enrolled": {
      "status": "pass",
      "message": "agent enrolled"
    },
    "certificates": {
      "status": "pass",
      "message": "certificate valid until 2026-12-16",
      "expires_in_days": 365
    },
    "hub_connection": {
      "status": "pass",
      "message": "last heartbeat 30s ago"
    },
    "policy": {
      "status": "pass",
      "message": "policy valid until 2026-12-31"
    },
    "disk_space": {
      "status": "pass",
      "message": "disk space 45.2% used"
    },
    "last_job": {
      "status": "pass",
      "message": "last job completed 5m ago"
    }
  },
  "version": "3.0.0",
  "agent_id": "agent-uuid"
}
```

**Unhealthy Response (503 Service Unavailable):**
```json
{
  "status": "unhealthy",
  "timestamp": "2025-12-16T10:30:00Z",
  "checks": {
    "hub_connection": {
      "status": "fail",
      "message": "no heartbeat for 600s"
    },
    "certificates": {
      "status": "warn",
      "message": "certificate expires in 15 days",
      "expires_in_days": 15
    }
  },
  "version": "3.0.0",
  "agent_id": "agent-uuid"
}
```

### Health Check Criteria

| Check | Pass | Warn | Fail |
|-------|------|------|------|
| Enrolled | Config and certs exist | - | Missing config/certs |
| Certificates | Expires in > 30 days | Expires in ≤ 30 days | Expired or invalid |
| Hub Connection | Last heartbeat < 5 min | - | Last heartbeat ≥ 5 min |
| Policy | Expires in > 7 days | Expires in ≤ 7 days | Expired |
| Disk Space | < 90% used | ≥ 90% used | - |
| Last Job | Success or no jobs | Last job failed | - |

### Using Health Checks

**curl:**
```bash
curl http://localhost:9091/health
```

**systemd integration:**
```ini
[Service]
ExecStartPost=/usr/bin/curl -f http://localhost:9091/health || exit 1
```

**Kubernetes liveness probe:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 9091
  initialDelaySeconds: 30
  periodSeconds: 10
```

## Certificate Management

### Automatic Renewal

Certificates are automatically renewed when:
- Expiration is within 30 days
- Daily check runs at 03:00 local time
- Manual renewal is triggered

### Manual Certificate Renewal

```bash
# Check certificate expiration
jtnt-agent cert check

# Renew certificate
jtnt-agent cert renew

# Rollback to previous certificate (if renewal fails)
jtnt-agent cert rollback
```

### Certificate Renewal Process

1. Agent checks certificate expiration daily
2. If < 30 days until expiry:
   - Generates Certificate Signing Request (CSR)
   - Sends renewal request to hub
   - Receives new certificate
   - Validates new certificate
   - Creates backup of current certificate
   - Atomically replaces certificate
   - Reloads mTLS client
3. Backup retained for 7 days

### Troubleshooting Certificate Issues

**Certificate expired:**
```bash
# Check expiration
openssl x509 -in /var/lib/jtnt-agent/certs/client.crt -noout -dates

# Force renewal
jtnt-agent cert renew --force
```

**Renewal failed:**
```bash
# Check logs
journalctl -u jtnt-agent | grep cert

# Check backup exists
ls -l /var/lib/jtnt-agent/certs/client.crt.backup

# Rollback if needed
jtnt-agent cert rollback
```

**Certificate chain validation errors:**
```bash
# Verify certificate chains to CA
openssl verify -CAfile /var/lib/jtnt-agent/certs/ca-bundle.crt \
  /var/lib/jtnt-agent/certs/client.crt
```

## Updates

### Checking for Updates

```bash
# Check if update is available
jtnt-agent update check

# Output:
# Update available: v3.1.0
# Current version: v3.0.0
# Release notes: Bug fixes and performance improvements
# Critical: false
```

### Applying Updates

**Manual update:**
```bash
# Download and apply update
jtnt-agent update apply

# Process:
# 1. Downloads new binary
# 2. Verifies SHA256 checksum
# 3. Verifies Ed25519 signature
# 4. Stops agent service
# 5. Backs up current binary
# 6. Installs new binary
# 7. Restarts service
# 8. Verifies new version running
```

**Automatic updates:**

Updates are checked daily at 04:00. Critical updates are applied automatically.

### Update Verification

All updates are verified with:
1. **SHA256 checksum**: Ensures download integrity
2. **Ed25519 signature**: Ensures authenticity

**Embedded public key** in binary verifies signatures.

### Rollback

If an update fails:

```bash
# Rollback to previous version
jtnt-agent update rollback
```

Automatic rollback occurs if:
- New binary fails to start
- Health check fails after update
- Service doesn't respond within 30 seconds

### Update Safety

- **Backup created**: Previous version saved as `.old`
- **Atomic replacement**: Binary swap is atomic
- **Service verification**: New version verified before cleanup
- **Automatic rollback**: On failure, previous version restored

## Troubleshooting

### Agent Won't Start

**Check service status:**
```bash
# systemd
sudo systemctl status jtnt-agent

# macOS
sudo launchctl list | grep jtnt

# Windows
sc query jtnt-agent
```

**Check logs:**
```bash
# systemd
journalctl -u jtnt-agent -n 100 --no-pager

# macOS
tail -f /var/log/jtnt-agent.log

# Windows
Get-EventLog -LogName Application -Source "JTNT Agent" -Newest 50
```

**Common issues:**
1. **Certificate missing**: Re-enroll agent
2. **Config invalid**: Check `/etc/jtnt-agent/config.json`
3. **Permissions**: Ensure agent runs as root/SYSTEM
4. **Port conflict**: Check if ports 9090/9091 are available

### Network Connectivity Issues

**Test hub connectivity:**
```bash
jtnt-agent test-connection
```

**Check network outage handling:**

Agent survives 72-hour network outages:
- Continues heartbeat attempts with exponential backoff
- Caches job results locally
- Automatically recovers when connectivity restored

**During outage:**
```bash
# Check cached results
ls -l ~/.jtnt/state/job_results_pending/

# Agent will automatically flush cache when connection restored
```

### High Resource Usage

**Check system metrics:**
```bash
# Via health endpoint
curl http://localhost:9091/health | jq '.checks.disk_space'

# Via metrics
curl http://localhost:9090/metrics | grep system_cpu
```

**Disk space cleanup:**
```bash
# Clean old job result cache
jtnt-agent cache clean

# Clean old audit logs (keeps 30 days by default)
jtnt-agent audit clean

# Clean update backups
jtnt-agent update clean
```

### Job Execution Failures

**Check job logs:**
```bash
# Recent jobs
jtnt-agent jobs list --recent 10

# Specific job
jtnt-agent jobs show <job-id>
```

**Policy violations:**
```bash
# View policy
jtnt-agent policy show

# Check if command is allowed
jtnt-agent policy check-exec /usr/bin/command

# Check if file access is allowed
jtnt-agent policy check-file read /path/to/file
```

### Circuit Breaker Open

If circuit breaker opens due to repeated failures:

```bash
# Check circuit breaker status
curl http://localhost:9090/metrics | grep circuit_breaker

# Wait for automatic recovery (1 minute timeout)
# Or restart agent to reset circuit breaker
sudo systemctl restart jtnt-agent
```

## Audit Logs

### Audit Log Location

```
/var/lib/jtnt-agent/audit/audit-YYYY-MM-DD.log
```

### Audit Log Format

Each entry is a JSON line with Ed25519 signature:

```json
{
  "timestamp": "2025-12-16T10:30:00Z",
  "type": "audit",
  "event": "job_executed",
  "agent_id": "agent-uuid",
  "job_id": "job-123",
  "command": "/usr/bin/systemctl status nginx",
  "status": "success",
  "user": "SYSTEM",
  "policy_version": 1,
  "details": {
    "exit_code": 0,
    "duration_ms": 234
  },
  "signature": "base64-ed25519-signature"
}
```

### Audit Events

- `job_executed`: Job execution completed
- `policy_changed`: Policy updated
- `cert_rotated`: Certificate renewed
- `update_applied`: Agent updated
- `enrollment`: Agent enrolled
- `policy_violation`: Policy rule violated
- `shutdown`: Agent shutdown
- `startup`: Agent started

### Viewing Audit Logs

```bash
# Today's audit log
cat /var/lib/jtnt-agent/audit/audit-$(date +%Y-%m-%d).log | jq .

# Search for job executions
grep job_executed /var/lib/jtnt-agent/audit/*.log | jq .

# Search for policy violations
grep policy_violation /var/lib/jtnt-agent/audit/*.log | jq .
```

### Verifying Audit Log Integrity

Audit logs are signed with agent's private key:

```bash
# Verify signature (requires agent public key)
jtnt-agent audit verify audit-2025-12-16.log
```

### Audit Log Retention

- Default retention: 30 days
- Automatic cleanup runs daily
- Configurable via config file

```json
{
  "audit_retention_days": 90
}
```

### Compliance

Audit logs provide tamper-evident trail for:
- SOC 2 compliance
- HIPAA audit requirements
- PCI DSS logging requirements
- General security auditing

Each entry is signed, making tampering detectable.

## Best Practices

### Monitoring

1. **Set up Prometheus** to scrape metrics
2. **Configure alerts** for critical metrics
3. **Monitor health endpoint** with your orchestrator
4. **Review audit logs** regularly

### Certificate Management

1. **Monitor expiration** via metrics
2. **Set alerts** for 30-day warning
3. **Test renewal** in staging first
4. **Keep backups** accessible

### Updates

1. **Test updates** in staging environment
2. **Schedule updates** during maintenance windows
3. **Monitor after update** for 24 hours
4. **Keep rollback capability** available

### Security

1. **Restrict metrics/health** to localhost only
2. **Protect audit logs** with appropriate permissions
3. **Review policy violations** promptly
4. **Rotate certificates** before expiration

### Performance

1. **Monitor resource usage** via metrics
2. **Clean caches** regularly
3. **Set appropriate** retry/backoff values
4. **Review slow jobs** in metrics
