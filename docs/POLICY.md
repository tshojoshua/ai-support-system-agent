# JTNT Agent Policy Reference

## Overview

The JTNT Agent uses a capability-based policy system to control what operations the agent can perform. All policies are signed with Ed25519 cryptographic signatures to prevent tampering.

## Policy Structure

```json
{
  "version": 1,
  "agent_id": "agent-123",
  "issued_at": 1704067200,
  "expires_at": 1735689600,
  "capabilities": {
    "exec": {
      "enabled": true,
      "allow_all": false,
      "binary_allowlist": [
        "/usr/bin/systemctl",
        "/usr/bin/apt",
        "/bin/ls"
      ]
    },
    "script": {
      "enabled": true,
      "allow_all": false,
      "interpreter_allowlist": [
        "/bin/bash",
        "/usr/bin/python3"
      ],
      "signature_required": true
    },
    "file": {
      "enabled": true,
      "allow_all": false,
      "read_allowlist": [
        "/var/log/**",
        "/etc/nginx/**"
      ],
      "write_allowlist": [
        "/tmp/**",
        "/var/app/config/**"
      ]
    }
  },
  "signature": "base64-encoded-ed25519-signature"
}
```

## Capability Types

### 1. Exec Capability

Controls execution of binary commands.

**Fields:**
- `enabled` (bool): Whether binary execution is allowed
- `allow_all` (bool): If true, any binary can be executed (use with caution)
- `binary_allowlist` ([]string): List of allowed binary paths (glob patterns supported)

**Path Matching:**
- Exact paths: `/usr/bin/systemctl`
- Wildcards: `/usr/bin/*`
- Recursive: `/usr/local/**`

**Example:**
```json
{
  "exec": {
    "enabled": true,
    "allow_all": false,
    "binary_allowlist": [
      "/usr/bin/apt",
      "/usr/bin/systemctl",
      "/bin/ps",
      "/usr/local/bin/**"
    ]
  }
}
```

### 2. Script Capability

Controls execution of interpreted scripts.

**Fields:**
- `enabled` (bool): Whether script execution is allowed
- `allow_all` (bool): If true, any interpreter can be used
- `interpreter_allowlist` ([]string): List of allowed interpreter paths
- `signature_required` (bool): If true, scripts must have valid Ed25519 signatures

**Script Signature Format:**
Scripts must be signed with Ed25519. The signature is passed separately in the job payload.

**Example:**
```json
{
  "script": {
    "enabled": true,
    "allow_all": false,
    "interpreter_allowlist": [
      "/bin/bash",
      "/bin/sh",
      "/usr/bin/python3",
      "/usr/bin/node"
    ],
    "signature_required": true
  }
}
```

### 3. File Capability

Controls file system read/write operations.

**Fields:**
- `enabled` (bool): Whether file operations are allowed
- `allow_all` (bool): If true, any file can be read/written
- `read_allowlist` ([]string): Paths that can be read
- `write_allowlist` ([]string): Paths that can be written

**Path Safety:**
- Symlink resolution prevents escaping allowed paths
- Parent directory traversal (`../`) is blocked
- Paths are canonicalized before matching

**Example:**
```json
{
  "file": {
    "enabled": true,
    "allow_all": false,
    "read_allowlist": [
      "/var/log/**",
      "/etc/nginx/**",
      "/proc/cpuinfo"
    ],
    "write_allowlist": [
      "/tmp/**",
      "/var/app/uploads/**",
      "/var/cache/app/**"
    ]
  }
}
```

## Default Policy

The agent starts with a secure default policy:

```go
policy.DefaultPolicy()
```

**Default Capabilities:**
- **Exec**: Disabled
- **Script**: Disabled
- **File**: Disabled

## Policy Distribution

Policies are distributed by the hub and cached locally:

1. Hub signs policy with private Ed25519 key
2. Agent downloads policy on startup or policy update event
3. Agent verifies signature with hub's public key
4. Valid policy is cached in `~/.jtnt/state/policy.json`
5. Agent reloads policy without restart on update

## Policy Enforcement

### Enforcement Points

1. **Job Acceptance**: Policy is checked before job execution starts
2. **Resource Access**: Checked before each file/binary/interpreter access
3. **Runtime**: Monitored throughout job execution

### Denial Behavior

When a job violates policy:
- Job immediately fails with `JobStatusFailed`
- Error message includes policy violation details
- Event is logged for audit trail
- No partial execution occurs

### Example Violation:

```json
{
  "status": "failed",
  "error": "policy violation: binary /bin/rm not in allowlist",
  "exit_code": -1
}
```

## Security Considerations

### 1. Signature Verification

All policies MUST be signed. Unsigned policies are rejected.

```go
if err := policy.VerifySignature(publicKey); err != nil {
    return fmt.Errorf("invalid policy signature: %w", err)
}
```

### 2. Path Canonicalization

All paths are canonicalized to prevent bypass:
- Symlinks are resolved
- Relative paths are converted to absolute
- `.` and `..` components are removed

### 3. Glob Pattern Safety

Glob patterns use secure matching:
- `*` matches single directory level
- `**` matches recursive directories
- No regex injection possible

### 4. Time-Based Expiration

Policies include expiration timestamps:
```json
{
  "expires_at": 1735689600
}
```

Expired policies are rejected automatically.

## Policy Updates

### Update Flow

1. Hub pushes policy update event
2. Agent fetches new signed policy
3. Signature is verified
4. New policy replaces current policy atomically
5. In-flight jobs continue with old policy
6. New jobs use new policy

### Rollback

If policy update fails verification:
- Current policy remains active
- Error is reported to hub
- Agent continues normal operation

## Audit Logging

All policy decisions are logged:

```json
{
  "event": "policy-check",
  "job_id": "job-123",
  "capability": "exec",
  "resource": "/usr/bin/systemctl",
  "allowed": true
}
```

## Best Practices

### 1. Principle of Least Privilege

Only enable capabilities needed for the agent's role:

```json
{
  "exec": {
    "enabled": true,
    "binary_allowlist": ["/usr/bin/systemctl"]
  },
  "script": {
    "enabled": false
  },
  "file": {
    "enabled": true,
    "read_allowlist": ["/var/log/nginx/**"]
  }
}
```

### 2. Use Specific Paths

Avoid wildcards when possible:
- ✅ Good: `"/usr/bin/apt"`
- ⚠️ Acceptable: `"/usr/local/bin/**"`
- ❌ Too broad: `"/**"`

### 3. Require Script Signatures

Always require signatures for production:

```json
{
  "script": {
    "signature_required": true
  }
}
```

### 4. Regular Policy Rotation

Update policies regularly:
- Set reasonable expiration times (e.g., 90 days)
- Rotate signing keys periodically
- Audit allowlists quarterly

## Testing Policies

Use the policy test tool:

```bash
# Validate policy structure
jtnt-agent policy validate policy.json

# Test binary execution
jtnt-agent policy check-exec /usr/bin/systemctl policy.json

# Test file access
jtnt-agent policy check-file read /var/log/nginx/access.log policy.json
```

## Example Policies

### Web Server Agent

```json
{
  "version": 1,
  "capabilities": {
    "exec": {
      "enabled": true,
      "binary_allowlist": [
        "/usr/bin/systemctl",
        "/usr/sbin/nginx"
      ]
    },
    "script": {
      "enabled": false
    },
    "file": {
      "enabled": true,
      "read_allowlist": ["/var/log/nginx/**", "/etc/nginx/**"],
      "write_allowlist": ["/etc/nginx/sites-available/**"]
    }
  }
}
```

### Database Agent

```json
{
  "version": 1,
  "capabilities": {
    "exec": {
      "enabled": true,
      "binary_allowlist": [
        "/usr/bin/psql",
        "/usr/bin/pg_dump"
      ]
    },
    "script": {
      "enabled": true,
      "interpreter_allowlist": ["/bin/bash"],
      "signature_required": true
    },
    "file": {
      "enabled": true,
      "read_allowlist": ["/var/lib/postgresql/**"],
      "write_allowlist": ["/var/backups/postgresql/**"]
    }
  }
}
```

### Monitoring Agent (Read-Only)

```json
{
  "version": 1,
  "capabilities": {
    "exec": {
      "enabled": true,
      "binary_allowlist": [
        "/bin/ps",
        "/usr/bin/top",
        "/usr/bin/free",
        "/usr/bin/df"
      ]
    },
    "script": {
      "enabled": false
    },
    "file": {
      "enabled": true,
      "read_allowlist": [
        "/proc/**",
        "/sys/**",
        "/var/log/**"
      ],
      "write_allowlist": []
    }
  }
}
```

## Troubleshooting

### Policy Rejected

**Error**: `invalid policy signature`

**Solution**: Ensure policy is signed with correct Ed25519 key.

### Binary Not Allowed

**Error**: `binary /usr/bin/curl not in allowlist`

**Solution**: Add binary to `binary_allowlist` or enable `allow_all` (not recommended for production).

### Path Not Matched

**Error**: `path /var/log/app/debug.log not in read allowlist`

**Solution**: Check glob patterns. Use `**` for recursive matching:
```json
{
  "read_allowlist": ["/var/log/**"]
}
```

### Policy Expired

**Error**: `policy expired at 2024-01-01T00:00:00Z`

**Solution**: Hub will push updated policy automatically. Check hub connectivity if policy is not updating.
