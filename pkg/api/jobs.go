package api

import (
	"time"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeExec     JobType = "exec"
	JobTypeScript   JobType = "script"
	JobTypeDownload JobType = "download"
	JobTypeUpload   JobType = "upload"
)

// Job represents a job to be executed
type Job struct {
	JobID      string                 `json:"job_id"`
	Type       JobType                `json:"type"`
	CreatedAt  time.Time              `json:"created_at"`
	TimeoutSec int                    `json:"timeout_sec"`
	Payload    map[string]interface{} `json:"payload"`
}

// ExecPayload represents exec job parameters
type ExecPayload struct {
	Binary     string   `json:"binary"`
	Args       []string `json:"args"`
	TimeoutSec int      `json:"timeout_sec"`
	WorkingDir string   `json:"working_dir"`
}

// ScriptPayload represents script job parameters
type ScriptPayload struct {
	Interpreter     string            `json:"interpreter"`
	ScriptContent   string            `json:"script_content"` // base64
	ScriptSignature string            `json:"script_signature"`
	TimeoutSec      int               `json:"timeout_sec"`
	EnvVars         map[string]string `json:"env_vars"`
}

// DownloadPayload represents download job parameters
type DownloadPayload struct {
	URL      string `json:"url"`
	DestPath string `json:"dest_path"`
	SHA256   string `json:"sha256"`
}

// UploadPayload represents upload job parameters
type UploadPayload struct {
	SourcePath   string `json:"source_path"`
	MaxSizeBytes int64  `json:"max_size_bytes"`
}

// JobStatus represents job execution status
type JobStatus string

const (
	StatusSuccess JobStatus = "success"
	StatusError   JobStatus = "error"
	StatusTimeout JobStatus = "timeout"
)

// JobResult represents the result of job execution
type JobResult struct {
	AgentID      string         `json:"agent_id"`
	Status       JobStatus      `json:"status"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   time.Time      `json:"finished_at"`
	ExitCode     int            `json:"exit_code,omitempty"`
	StdoutTail   string         `json:"stdout_tail,omitempty"` // base64, last 10KB
	StderrTail   string         `json:"stderr_tail,omitempty"` // base64, last 10KB
	ErrorMessage string         `json:"error_message,omitempty"`
	Artifacts    []ArtifactInfo `json:"artifacts,omitempty"`
}

// ArtifactInfo represents uploaded artifact metadata
type ArtifactInfo struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// UploadURL represents presigned upload URL
type UploadURL struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// ArtifactInitRequest requests upload URLs
type ArtifactInitRequest struct {
	JobID string         `json:"job_id"`
	Files []ArtifactInfo `json:"files"`
}

// ArtifactInitResponse contains upload URLs
type ArtifactInitResponse struct {
	UploadURLs []UploadURL `json:"upload_urls"`
}
