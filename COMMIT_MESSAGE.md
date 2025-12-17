# Git Commit Message for Phase 2

## Commit Title
feat: Implement Phase 2 - Secure Job Execution Engine

## Commit Body

Add comprehensive job execution system with capability-based policy enforcement.

### New Features

**Policy System:**
- Ed25519-signed policies with capability-based access control
- Policy enforcer with allowlist matching for binaries, scripts, and file paths
- Glob pattern support for flexible path matching
- Path canonicalization and symlink resolution for security

**Job Execution:**
- Four job types: exec, script, download, upload
- Exec handler: Binary execution without shell injection
- Script handler: Signed script execution with interpreter validation
- Download handler: File download with SHA256 verification
- Upload handler: Chunked file/directory uploads with presigned URLs

**Job Polling:**
- Automatic job polling from hub (default 30s interval)
- Exponential backoff on errors
- Context-based timeout enforcement
- Result caching and retry for failed uploads
- Automatic cache purge after 7 days

**Integration:**
- Job executor integrated into agent lifecycle
- Concurrent heartbeat and job polling loops
- Audit logging for all job executions
- Result buffering (last 10K lines or 1MB)

### Files Added (13)

Policy system:
- internal/policy/policy.go
- internal/policy/enforcer.go
- internal/policy/allowlist.go

Job execution:
- pkg/api/jobs.go
- internal/jobs/result.go
- internal/jobs/exec.go
- internal/jobs/script.go
- internal/jobs/download.go
- internal/jobs/upload.go
- internal/jobs/executor.go
- internal/jobs/executor_test.go

Agent integration:
- internal/agent/job_loop.go
- internal/agent/result_cache.go

### Files Modified (3)

- internal/agent/agent.go - Job executor integration
- internal/agent/logger.go - Add Audit method
- internal/transport/client.go - Add Upload method

### Documentation

- docs/POLICY.md - Comprehensive policy reference
- docs/PHASE2_SUMMARY.md - Implementation summary
- README.md - Updated with Phase 2 features

### Security

- All jobs enforced against signed policies
- No shell execution for binaries
- Script signature verification with Ed25519
- SHA256 hash verification for file downloads
- Secure temp file handling with 0700 permissions
- Output size limits to prevent DoS

### Testing

- Unit tests for job handlers
- Policy enforcement tests
- Result formatting tests

Closes #2 (if you have a GitHub issue for Phase 2)
