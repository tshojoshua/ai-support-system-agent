package jobs

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/policy"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

// ExecHandler executes binary commands
type ExecHandler struct {
	enforcer *policy.Enforcer
	agentID  string
}

// NewExecHandler creates a new exec handler
func NewExecHandler(enforcer *policy.Enforcer, agentID string) *ExecHandler {
	return &ExecHandler{
		enforcer: enforcer,
		agentID:  agentID,
	}
}

// Execute runs an exec job
func (h *ExecHandler) Execute(ctx context.Context, job *api.Job) *api.JobResult {
	startedAt := time.Now()

	// Parse payload
	var payload api.ExecPayload
	if err := ParsePayload(job.Payload, &payload); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("invalid payload: %w", err), nil)
	}

	// Determine timeout
	timeoutSec := payload.TimeoutSec
	if timeoutSec == 0 {
		timeoutSec = job.TimeoutSec
	}
	if timeoutSec == 0 {
		timeoutSec = h.enforcer.GetMaxExecTimeout()
	}

	// Enforce policy
	if err := h.enforcer.CanExecuteBinary(payload.Binary, timeoutSec); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("policy violation: %w", err), nil)
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Build command WITHOUT shell
	cmd := exec.CommandContext(execCtx, payload.Binary, payload.Args...)
	
	if payload.WorkingDir != "" {
		cmd.Dir = payload.WorkingDir
	}

	// Setup output capture
	stdout := NewTailBuffer(maxTailBytes)
	stderr := NewTailBuffer(maxTailBytes)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Execute
	err := cmd.Run()
	finishedAt := time.Now()

	// Determine status
	status := api.StatusSuccess
	exitCode := 0

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			status = api.StatusTimeout
		} else {
			status = api.StatusError
		}

		// Try to get exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return FormatResult(h.agentID, status, startedAt, finishedAt,
		exitCode, stdout, stderr, err, nil)
}
