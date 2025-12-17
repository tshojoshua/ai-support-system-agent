package jobs

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/policy"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

// mockClient implements basic transport.Client interface for testing
type mockClient struct {
	jobs    []*api.Job
	results map[string]*api.JobResult
}

func (m *mockClient) Get(ctx context.Context, path string) ([]byte, error) {
	return nil, nil
}

func (m *mockClient) Post(ctx context.Context, path string, data interface{}) ([]byte, error) {
	return nil, nil
}

// mockLogger for testing
type mockLogger struct{}

func (m *mockLogger) Info(event string, fields map[string]interface{})  {}
func (m *mockLogger) Error(event string, fields map[string]interface{}) {}
func (m *mockLogger) Debug(event string, fields map[string]interface{}) {}

func TestExecHandler_Execute(t *testing.T) {
	pol := policy.DefaultPolicy()
	pol.Capabilities.Exec.Enabled = true
	pol.Capabilities.Exec.AllowAll = true // Allow all for testing
	enforcer := policy.NewEnforcer(pol)

	handler := NewExecHandler(enforcer, &mockLogger{})

	tests := []struct {
		name       string
		job        *api.Job
		wantStatus api.JobStatus
	}{
		{
			name: "echo command",
			job: &api.Job{
				ID:      "test-1",
				Type:    api.JobTypeExec,
				Timeout: 5,
				Payload: map[string]interface{}{
					"binary": "echo",
					"args":   []interface{}{"hello", "world"},
				},
			},
			wantStatus: api.JobStatusCompleted,
		},
		{
			name: "pwd command",
			job: &api.Job{
				ID:      "test-2",
				Type:    api.JobTypeExec,
				Timeout: 5,
				Payload: map[string]interface{}{
					"binary": "pwd",
				},
			},
			wantStatus: api.JobStatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result := handler.Execute(ctx, tt.job)

			if result.Status != tt.wantStatus {
				t.Errorf("Execute() status = %v, want %v", result.Status, tt.wantStatus)
			}

			if result.Status == api.JobStatusCompleted && result.ExitCode != 0 {
				t.Errorf("Execute() exitCode = %v, want 0", result.ExitCode)
			}

			if result.Output == "" {
				t.Errorf("Execute() output is empty")
			}

			t.Logf("Output: %s", result.Output)
		})
	}
}

func TestExecHandler_PolicyEnforcement(t *testing.T) {
	pol := policy.DefaultPolicy()
	pol.Capabilities.Exec.Enabled = true
	pol.Capabilities.Exec.AllowAll = false
	pol.Capabilities.Exec.BinaryAllowlist = []string{"/bin/echo"}
	
	enforcer := policy.NewEnforcer(pol)
	handler := NewExecHandler(enforcer, &mockLogger{})

	tests := []struct {
		name       string
		binary     string
		wantStatus api.JobStatus
	}{
		{
			name:       "allowed binary",
			binary:     "/bin/echo",
			wantStatus: api.JobStatusCompleted,
		},
		{
			name:       "denied binary",
			binary:     "/bin/rm",
			wantStatus: api.JobStatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			job := &api.Job{
				ID:      "test-policy",
				Type:    api.JobTypeExec,
				Timeout: 5,
				Payload: map[string]interface{}{
					"binary": tt.binary,
					"args":   []interface{}{"test"},
				},
			}

			result := handler.Execute(ctx, job)

			if result.Status != tt.wantStatus {
				t.Errorf("Execute() status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantStatus == api.JobStatusFailed {
				if !strings.Contains(result.Error, "policy") {
					t.Errorf("Expected policy error, got: %s", result.Error)
				}
			}
		})
	}
}

func TestResultFormatting(t *testing.T) {
	result := NewJobResult()

	// Test tail buffer
	for i := 0; i < 2000; i++ {
		result.AppendOutput("Line %d\n", i)
	}

	output := result.Finalize()

	// Output should be truncated
	if len(output) > maxOutputBytes {
		t.Errorf("Output exceeds max bytes: %d > %d", len(output), maxOutputBytes)
	}

	// Should contain truncation notice
	if !strings.Contains(output, "[truncated]") {
		t.Errorf("Output should contain truncation notice")
	}

	// Should contain last lines
	if !strings.Contains(output, "Line 1999") {
		t.Errorf("Output should contain last line")
	}
}

func TestScriptExecution_Basic(t *testing.T) {
	pol := policy.DefaultPolicy()
	pol.Capabilities.Script.Enabled = true
	pol.Capabilities.Script.AllowAll = true
	pol.Capabilities.Script.SignatureRequired = false // Disable for basic test
	
	enforcer := policy.NewEnforcer(pol)
	handler := NewScriptHandler(enforcer, &mockLogger{})

	job := &api.Job{
		ID:      "test-script",
		Type:    api.JobTypeScript,
		Timeout: 5,
		Payload: map[string]interface{}{
			"interpreter": "/bin/bash",
			"script":      "#!/bin/bash\necho 'Hello from script'\n",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := handler.Execute(ctx, job)

	if result.Status != api.JobStatusCompleted {
		t.Errorf("Script execution failed: %s", result.Error)
	}

	if !strings.Contains(result.Output, "Hello from script") {
		t.Errorf("Script output not found, got: %s", result.Output)
	}
}
