# Testing Guide

## Running Tests

### Prerequisites

Ensure Go 1.23+ is installed:

```bash
go version
```

### Run All Tests

```bash
# Using Makefile
make test

# Or directly with Go
go test ./... -v
```

### Run Specific Package Tests

```bash
# Config tests
go test ./internal/config -v

# Store tests
go test ./internal/store -v

# Enrollment tests
go test ./internal/enroll -v

# Transport/retry tests
go test ./internal/transport -v

# System info tests
go test ./internal/sysinfo -v
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
open coverage.html
```

### Race Condition Detection

```bash
go test ./... -race
```

## Test Structure

### Unit Tests Included

1. **Config Tests** (`internal/config/config_test.go`)
   - Configuration validation
   - Save/load functionality
   - Error handling

2. **Store Tests** (`internal/store/store_test.go`)
   - Secure file storage
   - Permission validation
   - CRUD operations

3. **Keypair Tests** (`internal/enroll/keypair_test.go`)
   - Ed25519 key generation
   - Base64 encoding/decoding
   - Key validation

4. **Retry Logic Tests** (`internal/transport/retry_test.go`)
   - Exponential backoff calculation
   - Error classification (retryable vs non-retryable)
   - Context cancellation
   - Max attempts enforcement

5. **System Info Tests** (`internal/sysinfo/sysinfo_test.go`)
   - Metric collection
   - Data validation
   - Performance benchmarks

## Manual Testing

### 1. Build the Agent

```bash
make build
```

### 2. Test Enrollment (requires test hub)

```bash
# Set up test hub or use mock server
./bin/jtnt-agent enroll --token TEST_TOKEN --hub http://localhost:8080
```

### 3. Test Status Command

```bash
./bin/jtnt-agent status
```

### 4. Test Version Command

```bash
./bin/jtnt-agent version
```

### 5. Run Agent Manually

```bash
sudo ./bin/jtnt-agentd
```

## Integration Testing (Future)

Phase 1 focuses on unit tests. Integration tests will be added in future phases:

- End-to-end enrollment with real hub
- Heartbeat transmission and verification
- Certificate rotation
- Service installation and lifecycle

## Continuous Integration

Recommended CI pipeline:

```yaml
# Example GitHub Actions workflow
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.23']
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go }}
      
      - name: Install dependencies
        run: make deps
      
      - name: Run tests
        run: make test
      
      - name: Build
        run: make build
```

## Performance Testing

### Benchmark System Info Collection

```bash
go test ./internal/sysinfo -bench=. -benchmem
```

Expected results:
- ~1-2ms per collection
- Minimal memory allocation

### Stress Test Heartbeat Loop

Create a test that runs heartbeat loop for extended period:

```go
// Future: Add to integration tests
func TestHeartbeatStress(t *testing.T) {
    // Run heartbeat for 1 hour
    // Monitor memory usage
    // Verify no goroutine leaks
}
```

## Platform-Specific Testing

### Linux

```bash
# Test on Ubuntu 20.04, 22.04
# Test on Debian 11, 12
make build-linux
```

### macOS

```bash
# Test on macOS 13+ (Intel)
GOARCH=amd64 make build-darwin

# Test on macOS 13+ (Apple Silicon)
GOARCH=arm64 make build-darwin
```

### Windows

```bash
# Test on Windows 10, 11, Server 2019/2022
make build-windows
```

## Security Testing

### Certificate Validation

```bash
# Verify certificate permissions
ls -la /var/lib/jtnt-agent/certs/

# Should be 0600 (owner read/write only)
```

### Secure Storage

```bash
# Test permission enforcement
go test ./internal/store -v -run TestStore_Permissions
```

## Debugging Tests

### Verbose Output

```bash
go test ./... -v
```

### Run Specific Test

```bash
go test ./internal/config -run TestConfigValidation -v
```

### Debug with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/config -- -test.run TestConfigValidation
```

## Test Maintenance

### Adding New Tests

1. Create `*_test.go` file in same package
2. Follow naming convention: `Test<FunctionName>`
3. Use table-driven tests for multiple cases
4. Include error cases
5. Add benchmarks for performance-critical code

### Test Coverage Goals

- Unit test coverage: >80%
- Critical paths: 100%
- Error handling: Complete coverage

## Known Test Limitations

Phase 1 tests do not cover:

- Network communication (mocked in unit tests)
- Actual hub enrollment (requires test infrastructure)
- Service installation
- Long-running stability tests

These will be addressed in future phases.
