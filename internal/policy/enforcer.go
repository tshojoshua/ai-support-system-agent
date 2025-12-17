package policy

import (
	"errors"
	"fmt"
)

var (
	// ErrPolicyExpired indicates policy has expired
	ErrPolicyExpired = errors.New("policy expired")

	// ErrCapabilityDisabled indicates capability is not enabled
	ErrCapabilityDisabled = errors.New("capability disabled")

	// ErrBinaryNotAllowed indicates binary is not in allowlist
	ErrBinaryNotAllowed = errors.New("binary not allowed")

	// ErrPathNotAllowed indicates path is not in allowlist
	ErrPathNotAllowed = errors.New("path not allowed")

	// ErrTimeoutExceeded indicates timeout exceeds policy maximum
	ErrTimeoutExceeded = errors.New("timeout exceeds policy maximum")

	// ErrInterpreterNotAllowed indicates script interpreter is not allowed
	ErrInterpreterNotAllowed = errors.New("interpreter not allowed")

	// ErrSignatureRequired indicates script signature is required but missing
	ErrSignatureRequired = errors.New("script signature required")

	// ErrFileSizeExceeded indicates file size exceeds policy limit
	ErrFileSizeExceeded = errors.New("file size exceeds policy limit")

	// ErrPathTraversal indicates path contains traversal attempt
	ErrPathTraversal = errors.New("path traversal detected")
)

// Enforcer enforces policy rules
type Enforcer struct {
	policy *Policy
}

// NewEnforcer creates a new policy enforcer
func NewEnforcer(policy *Policy) (*Enforcer, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}

	return &Enforcer{policy: policy}, nil
}

// CanExecuteBinary checks if binary execution is allowed
func (e *Enforcer) CanExecuteBinary(binary string, timeoutSec int) error {
	if e.policy.Capabilities.Exec == nil || !e.policy.Capabilities.Exec.Enabled {
		return ErrCapabilityDisabled
	}

	exec := e.policy.Capabilities.Exec

	// Check binary allowlist
	if !AllowsBinary(exec.AllowedBinaries, binary) {
		// Check if binary path is in allowed paths
		pathAllowlist := NewAllowlist(exec.AllowedPaths)
		if !pathAllowlist.Allows(binary) {
			return fmt.Errorf("%w: %s", ErrBinaryNotAllowed, binary)
		}
	}

	// Check timeout
	if timeoutSec > exec.MaxExecutionSec {
		return fmt.Errorf("%w: %d > %d", ErrTimeoutExceeded, timeoutSec, exec.MaxExecutionSec)
	}

	return nil
}

// CanExecuteScript checks if script execution is allowed
func (e *Enforcer) CanExecuteScript(interpreter string, scriptSize int, hasSignature bool, timeoutSec int) error {
	if e.policy.Capabilities.Script == nil || !e.policy.Capabilities.Script.Enabled {
		return ErrCapabilityDisabled
	}

	script := e.policy.Capabilities.Script

	// Check interpreter
	interpreterAllowed := false
	for _, allowed := range script.AllowedInterpreters {
		if interpreter == allowed {
			interpreterAllowed = true
			break
		}
	}
	if !interpreterAllowed {
		return fmt.Errorf("%w: %s", ErrInterpreterNotAllowed, interpreter)
	}

	// Check signature requirement
	if script.RequireSignature && !hasSignature {
		return ErrSignatureRequired
	}

	// Check script size
	if scriptSize > script.MaxScriptSizeBytes {
		return fmt.Errorf("%w: %d > %d", ErrFileSizeExceeded, scriptSize, script.MaxScriptSizeBytes)
	}

	// Check timeout
	if timeoutSec > script.MaxExecutionSec {
		return fmt.Errorf("%w: %d > %d", ErrTimeoutExceeded, timeoutSec, script.MaxExecutionSec)
	}

	return nil
}

// CanReadFile checks if file read is allowed
func (e *Enforcer) CanReadFile(path string) error {
	if e.policy.Capabilities.File == nil {
		return ErrCapabilityDisabled
	}

	// Validate path for traversal
	if err := ValidatePath(path); err != nil {
		return err
	}

	allowlist := NewAllowlist(e.policy.Capabilities.File.ReadPaths)
	if !allowlist.Allows(path) {
		return fmt.Errorf("%w: %s", ErrPathNotAllowed, path)
	}

	return nil
}

// CanWriteFile checks if file write is allowed
func (e *Enforcer) CanWriteFile(path string, size int64) error {
	if e.policy.Capabilities.File == nil {
		return ErrCapabilityDisabled
	}

	// Validate path for traversal
	if err := ValidatePath(path); err != nil {
		return err
	}

	allowlist := NewAllowlist(e.policy.Capabilities.File.WritePaths)
	if !allowlist.Allows(path) {
		return fmt.Errorf("%w: %s", ErrPathNotAllowed, path)
	}

	// Check file size
	if size > e.policy.Capabilities.File.MaxFileSizeBytes {
		return fmt.Errorf("%w: %d > %d", ErrFileSizeExceeded, size, e.policy.Capabilities.File.MaxFileSizeBytes)
	}

	return nil
}

// GetMaxExecTimeout returns maximum execution timeout
func (e *Enforcer) GetMaxExecTimeout() int {
	if e.policy.Capabilities.Exec != nil {
		return e.policy.Capabilities.Exec.MaxExecutionSec
	}
	return 300 // Default 5 minutes
}

// GetMaxScriptTimeout returns maximum script timeout
func (e *Enforcer) GetMaxScriptTimeout() int {
	if e.policy.Capabilities.Script != nil {
		return e.policy.Capabilities.Script.MaxExecutionSec
	}
	return 600 // Default 10 minutes
}

// Policy returns the underlying policy
func (e *Enforcer) Policy() *Policy {
	return e.policy
}
