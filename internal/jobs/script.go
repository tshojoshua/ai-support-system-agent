package jobs

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/policy"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

// ScriptHandler executes scripts
type ScriptHandler struct {
	enforcer  *policy.Enforcer
	agentID   string
	publicKey ed25519.PublicKey
}

// NewScriptHandler creates a new script handler
func NewScriptHandler(enforcer *policy.Enforcer, agentID string, publicKey ed25519.PublicKey) *ScriptHandler {
	return &ScriptHandler{
		enforcer:  enforcer,
		agentID:   agentID,
		publicKey: publicKey,
	}
}

// Execute runs a script job
func (h *ScriptHandler) Execute(ctx context.Context, job *api.Job) *api.JobResult {
	startedAt := time.Now()

	// Parse payload
	var payload api.ScriptPayload
	if err := ParsePayload(job.Payload, &payload); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("invalid payload: %w", err), nil)
	}

	// Decode script content
	scriptBytes, err := base64.StdEncoding.DecodeString(payload.ScriptContent)
	if err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("invalid script encoding: %w", err), nil)
	}

	// Determine timeout
	timeoutSec := payload.TimeoutSec
	if timeoutSec == 0 {
		timeoutSec = job.TimeoutSec
	}
	if timeoutSec == 0 {
		timeoutSec = h.enforcer.GetMaxScriptTimeout()
	}

	// Enforce policy
	hasSignature := payload.ScriptSignature != ""
	if err := h.enforcer.CanExecuteScript(payload.Interpreter, len(scriptBytes), hasSignature, timeoutSec); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("policy violation: %w", err), nil)
	}

	// Verify signature if provided
	if hasSignature {
		if err := h.verifyScriptSignature(scriptBytes, payload.ScriptSignature); err != nil {
			return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
				-1, nil, nil, fmt.Errorf("signature verification failed: %w", err), nil)
		}
	}

	// Create temp script file
	scriptPath, cleanup, err := h.createTempScript(scriptBytes, payload.Interpreter)
	if err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("failed to create script file: %w", err), nil)
	}
	defer cleanup()

	// Execute script
	return h.executeScript(ctx, scriptPath, payload.Interpreter, payload.EnvVars, timeoutSec, startedAt)
}

func (h *ScriptHandler) verifyScriptSignature(script []byte, signatureB64 string) error {
	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	if !ed25519.Verify(h.publicKey, script, sig) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

func (h *ScriptHandler) createTempScript(content []byte, interpreter string) (string, func(), error) {
	// Determine extension
	ext := ".sh"
	switch interpreter {
	case "powershell":
		ext = ".ps1"
	case "bash", "sh":
		ext = ".sh"
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "jtnt-script-*"+ext)
	if err != nil {
		return "", nil, err
	}

	scriptPath := tmpFile.Name()

	// Write content
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		os.Remove(scriptPath)
		return "", nil, err
	}

	// Set permissions (owner read/execute only)
	if err := tmpFile.Chmod(0700); err != nil {
		tmpFile.Close()
		os.Remove(scriptPath)
		return "", nil, err
	}

	tmpFile.Close()

	cleanup := func() {
		os.Remove(scriptPath)
	}

	return scriptPath, cleanup, nil
}

func (h *ScriptHandler) executeScript(ctx context.Context, scriptPath, interpreter string,
	envVars map[string]string, timeoutSec int, startedAt time.Time) *api.JobResult {

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Build command based on interpreter
	var cmd *exec.Cmd
	switch interpreter {
	case "powershell":
		cmd = exec.CommandContext(execCtx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	case "bash":
		cmd = exec.CommandContext(execCtx, "bash", scriptPath)
	case "sh":
		cmd = exec.CommandContext(execCtx, "sh", scriptPath)
	default:
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("unsupported interpreter: %s", interpreter), nil)
	}

	// Set environment variables
	if len(envVars) > 0 {
		cmd.Env = os.Environ()
		for k, v := range envVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
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

		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return FormatResult(h.agentID, status, startedAt, finishedAt,
		exitCode, stdout, stderr, err, nil)
}
