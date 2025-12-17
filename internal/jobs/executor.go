package jobs

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"

	"github.com/tshojoshua/jtnt-agent/internal/policy"
	"github.com/tshojoshua/jtnt-agent/internal/transport"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

// Executor orchestrates job execution
type Executor struct {
	agentID       string
	enforcer      *policy.Enforcer
	client        *transport.Client
	publicKey     ed25519.PublicKey
	execHandler   *ExecHandler
	scriptHandler *ScriptHandler
	downloadHandler *DownloadHandler
	uploadHandler *UploadHandler
	logger        JobLogger
}

// JobLogger interface for logging job execution
type JobLogger interface {
	Info(component string, fields map[string]interface{})
	Error(component string, fields map[string]interface{})
	Audit(jobID string, jobType string, status string, fields map[string]interface{})
}

// NewExecutor creates a new job executor
func NewExecutor(agentID string, enforcer *policy.Enforcer, client *transport.Client, 
	publicKey ed25519.PublicKey, logger JobLogger) *Executor {
	
	exec := &Executor{
		agentID:   agentID,
		enforcer:  enforcer,
		client:    client,
		publicKey: publicKey,
		logger:    logger,
	}

	// Initialize handlers
	exec.execHandler = NewExecHandler(enforcer, agentID)
	exec.scriptHandler = NewScriptHandler(enforcer, agentID, publicKey)
	exec.downloadHandler = NewDownloadHandler(enforcer, agentID)
	exec.uploadHandler = NewUploadHandler(enforcer, agentID, client)

	return exec
}

// Execute executes a job based on its type
func (e *Executor) Execute(ctx context.Context, job *api.Job) *api.JobResult {
	e.logger.Info("job", map[string]interface{}{
		"message": "executing job",
		"job_id":  job.JobID,
		"type":    job.Type,
	})

	var result *api.JobResult

	switch job.Type {
	case api.JobTypeExec:
		result = e.execHandler.Execute(ctx, job)
	case api.JobTypeScript:
		result = e.scriptHandler.Execute(ctx, job)
	case api.JobTypeDownload:
		result = e.downloadHandler.Execute(ctx, job)
	case api.JobTypeUpload:
		result = e.uploadHandler.Execute(ctx, job)
	default:
		result = &api.JobResult{
			AgentID:      e.agentID,
			Status:       api.StatusError,
			ErrorMessage: fmt.Sprintf("unsupported job type: %s", job.Type),
		}
	}

	// Audit log
	e.logger.Audit(job.JobID, string(job.Type), string(result.Status), map[string]interface{}{
		"exit_code":      result.ExitCode,
		"error_message":  result.ErrorMessage,
		"policy_version": e.enforcer.Policy().Version,
	})

	return result
}

// FetchNextJob fetches the next pending job from hub
func (e *Executor) FetchNextJob(ctx context.Context) (*api.Job, error) {
	respData, err := e.client.Get(ctx, "/api/v1/agent/jobs/next")
	if err != nil {
		// 204 No Content means no jobs available
		if err.Error() == "unexpected status code: 204" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	var job api.Job
	if err := json.Unmarshal(respData, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job: %w", err)
	}

	return &job, nil
}

// ReportResult reports job result to hub
func (e *Executor) ReportResult(ctx context.Context, jobID string, result *api.JobResult) error {
	path := fmt.Sprintf("/api/v1/agent/jobs/%s/result", jobID)
	
	_, err := e.client.Post(ctx, path, result)
	if err != nil {
		return fmt.Errorf("failed to report result: %w", err)
	}

	return nil
}
