package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/policy"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

// DownloadHandler handles file downloads
type DownloadHandler struct {
	enforcer *policy.Enforcer
	agentID  string
}

// NewDownloadHandler creates a new download handler
func NewDownloadHandler(enforcer *policy.Enforcer, agentID string) *DownloadHandler {
	return &DownloadHandler{
		enforcer: enforcer,
		agentID:  agentID,
	}
}

// Execute downloads a file
func (h *DownloadHandler) Execute(ctx context.Context, job *api.Job) *api.JobResult {
	startedAt := time.Now()

	// Parse payload
	var payload api.DownloadPayload
	if err := ParsePayload(job.Payload, &payload); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("invalid payload: %w", err), nil)
	}

	// Enforce policy - check write permission
	if err := h.enforcer.CanWriteFile(payload.DestPath, 0); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("policy violation: %w", err), nil)
	}

	// Ensure directory exists
	dir := filepath.Dir(payload.DestPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("failed to create directory: %w", err), nil)
	}

	// Download file
	if err := h.downloadFile(ctx, payload.URL, payload.DestPath, payload.SHA256); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, err, nil)
	}

	finishedAt := time.Now()
	return FormatResult(h.agentID, api.StatusSuccess, startedAt, finishedAt,
		0, nil, nil, nil, nil)
}

func (h *DownloadHandler) downloadFile(ctx context.Context, url, destPath, expectedHash string) error {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tempPath := destPath + ".tmp"
	outFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		outFile.Close()
		os.Remove(tempPath) // Clean up temp file if still exists
	}()

	// Download with hash calculation
	hasher := sha256.New()
	writer := io.MultiWriter(outFile, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	outFile.Close()

	// Verify hash if provided
	if expectedHash != "" {
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if actualHash != expectedHash {
			return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
		}
	}

	// Move temp file to final destination
	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Set permissions
	if err := os.Chmod(destPath, 0600); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}
