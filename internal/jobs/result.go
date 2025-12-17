package jobs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

const (
	maxTailBytes = 10 * 1024 // 10KB
)

// TailBuffer captures the last N bytes of output
type TailBuffer struct {
	buf     *bytes.Buffer
	maxSize int
}

// NewTailBuffer creates a new tail buffer
func NewTailBuffer(maxSize int) *TailBuffer {
	return &TailBuffer{
		buf:     &bytes.Buffer{},
		maxSize: maxSize,
	}
}

// Write implements io.Writer
func (t *TailBuffer) Write(p []byte) (n int, err error) {
	n = len(p)
	
	// Write to buffer
	t.buf.Write(p)
	
	// Keep only last maxSize bytes
	if t.buf.Len() > t.maxSize {
		// Create new buffer with last maxSize bytes
		data := t.buf.Bytes()
		offset := len(data) - t.maxSize
		t.buf = bytes.NewBuffer(data[offset:])
	}
	
	return n, nil
}

// Bytes returns the buffered bytes
func (t *TailBuffer) Bytes() []byte {
	return t.buf.Bytes()
}

// Base64 returns base64-encoded buffer content
func (t *TailBuffer) Base64() string {
	if t.buf.Len() == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(t.buf.Bytes())
}

// FormatResult creates a JobResult from execution details
func FormatResult(agentID string, status api.JobStatus, startedAt, finishedAt time.Time,
	exitCode int, stdout, stderr *TailBuffer, err error, artifacts []api.ArtifactInfo) *api.JobResult {
	
	result := &api.JobResult{
		AgentID:    agentID,
		Status:     status,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		ExitCode:   exitCode,
		Artifacts:  artifacts,
	}

	if stdout != nil {
		result.StdoutTail = stdout.Base64()
	}

	if stderr != nil {
		result.StderrTail = stderr.Base64()
	}

	if err != nil {
		result.ErrorMessage = err.Error()
	}

	return result
}

// ParsePayload parses job payload into specific type
func ParsePayload(payload map[string]interface{}, target interface{}) error {
	// Convert map to JSON and back to struct
	// This is a simple approach; production code might use mapstructure
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}
