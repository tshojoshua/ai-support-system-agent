package jobs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/policy"
	"github.com/tshojoshua/jtnt-agent/internal/transport"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

const (
	chunkSize = 5 * 1024 * 1024 // 5MB chunks
)

// UploadHandler handles file uploads
type UploadHandler struct {
	enforcer *policy.Enforcer
	agentID  string
	client   *transport.Client
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(enforcer *policy.Enforcer, agentID string, client *transport.Client) *UploadHandler {
	return &UploadHandler{
		enforcer: enforcer,
		agentID:  agentID,
		client:   client,
	}
}

// Execute uploads a file
func (h *UploadHandler) Execute(ctx context.Context, job *api.Job) *api.JobResult {
	startedAt := time.Now()

	// Parse payload
	var payload api.UploadPayload
	if err := ParsePayload(job.Payload, &payload); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("invalid payload: %w", err), nil)
	}

	// Enforce policy - check read permission
	if err := h.enforcer.CanReadFile(payload.SourcePath); err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("policy violation: %w", err), nil)
	}

	// Get file info
	fileInfo, err := os.Stat(payload.SourcePath)
	if err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("failed to stat file: %w", err), nil)
	}

	// Check file size
	maxSize := payload.MaxSizeBytes
	if maxSize == 0 {
		maxSize = h.enforcer.Policy().Capabilities.File.MaxFileSizeBytes
	}
	if fileInfo.Size() > maxSize {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, fmt.Errorf("file size %d exceeds maximum %d", fileInfo.Size(), maxSize), nil)
	}

	// Upload file(s)
	artifacts, err := h.uploadPath(ctx, job.JobID, payload.SourcePath)
	if err != nil {
		return FormatResult(h.agentID, api.StatusError, startedAt, time.Now(),
			-1, nil, nil, err, nil)
	}

	finishedAt := time.Now()
	return FormatResult(h.agentID, api.StatusSuccess, startedAt, finishedAt,
		0, nil, nil, nil, artifacts)
}

func (h *UploadHandler) uploadPath(ctx context.Context, jobID, path string) ([]api.ArtifactInfo, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var artifacts []api.ArtifactInfo

	if fileInfo.IsDir() {
		// Upload directory (recursive)
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				artifact, err := h.uploadFile(ctx, jobID, filePath)
				if err != nil {
					return err
				}
				artifacts = append(artifacts, *artifact)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Upload single file
		artifact, err := h.uploadFile(ctx, jobID, path)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, *artifact)
	}

	return artifacts, nil
}

func (h *UploadHandler) uploadFile(ctx context.Context, jobID, filePath string) (*api.ArtifactInfo, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate SHA256
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to hash file: %w", err)
	}
	sha256Hash := hex.EncodeToString(hasher.Sum(nil))

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	// Create artifact info
	artifact := api.ArtifactInfo{
		Name:   filepath.Base(filePath),
		Size:   fileInfo.Size(),
		SHA256: sha256Hash,
	}

	// Initialize upload
	uploadURL, err := h.initializeUpload(ctx, jobID, []api.ArtifactInfo{artifact})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize upload: %w", err)
	}

	if len(uploadURL) == 0 {
		return nil, fmt.Errorf("no upload URL received")
	}

	// Upload file to presigned URL
	if err := h.uploadToURL(ctx, uploadURL[0], file, fileInfo.Size()); err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}

	return &artifact, nil
}

func (h *UploadHandler) initializeUpload(ctx context.Context, jobID string, artifacts []api.ArtifactInfo) ([]api.UploadURL, error) {
	req := api.ArtifactInitRequest{
		JobID: jobID,
		Files: artifacts,
	}

	respData, err := h.client.Post(ctx, "/api/v1/agent/artifacts/init", req)
	if err != nil {
		return nil, err
	}

	var resp api.ArtifactInitResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.UploadURLs, nil
}

func (h *UploadHandler) uploadToURL(ctx context.Context, uploadURL api.UploadURL, reader io.Reader, size int64) error {
	// Create request
	req, err := http.NewRequestWithContext(ctx, uploadURL.Method, uploadURL.URL, reader)
	if err != nil {
		return err
	}

	// Set headers
	for k, v := range uploadURL.Headers {
		req.Header.Set(k, v)
	}
	req.ContentLength = size

	// Execute upload
	client := &http.Client{
		Timeout: 30 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
