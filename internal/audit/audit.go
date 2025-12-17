package audit

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/config"
)

// EventType represents the type of audit event
type EventType string

const (
	EventJobExecuted       EventType = "job_executed"
	EventPolicyChanged     EventType = "policy_changed"
	EventCertRotated       EventType = "cert_rotated"
	EventUpdateApplied     EventType = "update_applied"
	EventEnrollment        EventType = "enrollment"
	EventPolicyViolation   EventType = "policy_violation"
	EventShutdown          EventType = "shutdown"
	EventStartup           EventType = "startup"
)

// Entry represents a single audit log entry
type Entry struct {
	Timestamp     string                 `json:"timestamp"`
	Type          string                 `json:"type"`
	Event         string                 `json:"event"`
	AgentID       string                 `json:"agent_id,omitempty"`
	JobID         string                 `json:"job_id,omitempty"`
	Command       string                 `json:"command,omitempty"`
	Status        string                 `json:"status,omitempty"`
	User          string                 `json:"user,omitempty"`
	PolicyVersion int                    `json:"policy_version,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Signature     string                 `json:"signature"`
}

// Logger handles audit logging with signatures
type Logger struct {
	mu         sync.Mutex
	file       *os.File
	filePath   string
	privateKey ed25519.PrivateKey
	agentID    string
}

// NewLogger creates a new audit logger
func NewLogger(agentID string, privateKey ed25519.PrivateKey) (*Logger, error) {
	auditDir := filepath.Join(config.GetStateDir(), "audit")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %w", err)
	}

	// Create audit log file with date
	filename := fmt.Sprintf("audit-%s.log", time.Now().Format("2006-01-02"))
	filePath := filepath.Join(auditDir, filename)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}

	return &Logger{
		file:       file,
		filePath:   filePath,
		privateKey: privateKey,
		agentID:    agentID,
	}, nil
}

// Log writes an audit entry
func (l *Logger) Log(event EventType, details map[string]interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create entry without signature first
	entry := &Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Type:      "audit",
		Event:     string(event),
		AgentID:   l.agentID,
		Details:   details,
	}

	// Extract common fields from details
	if jobID, ok := details["job_id"].(string); ok {
		entry.JobID = jobID
	}
	if command, ok := details["command"].(string); ok {
		entry.Command = command
	}
	if status, ok := details["status"].(string); ok {
		entry.Status = status
	}
	if policyVersion, ok := details["policy_version"].(int); ok {
		entry.PolicyVersion = policyVersion
	}

	// Set user (system by default)
	entry.User = "SYSTEM"
	if user, ok := details["user"].(string); ok {
		entry.User = user
	}

	// Generate signature
	signature, err := l.signEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to sign entry: %w", err)
	}
	entry.Signature = signature

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Write to file
	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	// Sync to disk
	return l.file.Sync()
}

// signEntry creates a signature for an audit entry
func (l *Logger) signEntry(entry *Entry) (string, error) {
	// Create canonical representation for signing (without signature field)
	data, err := json.Marshal(map[string]interface{}{
		"timestamp":      entry.Timestamp,
		"type":           entry.Type,
		"event":          entry.Event,
		"agent_id":       entry.AgentID,
		"job_id":         entry.JobID,
		"command":        entry.Command,
		"status":         entry.Status,
		"user":           entry.User,
		"policy_version": entry.PolicyVersion,
		"details":        entry.Details,
	})
	if err != nil {
		return "", err
	}

	// Sign the data
	signature := ed25519.Sign(l.privateKey, data)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifyEntry verifies the signature of an audit entry
func VerifyEntry(entry *Entry, publicKey ed25519.PublicKey) error {
	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(entry.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Create canonical data (same as signing)
	data, err := json.Marshal(map[string]interface{}{
		"timestamp":      entry.Timestamp,
		"type":           entry.Type,
		"event":          entry.Event,
		"agent_id":       entry.AgentID,
		"job_id":         entry.JobID,
		"command":        entry.Command,
		"status":         entry.Status,
		"user":           entry.User,
		"policy_version": entry.PolicyVersion,
		"details":        entry.Details,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Verify signature
	if !ed25519.Verify(publicKey, data, signature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// Close closes the audit log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Rotate rotates the audit log file
func (l *Logger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Close current file
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return err
		}
	}

	// Open new file with current date
	filename := fmt.Sprintf("audit-%s.log", time.Now().Format("2006-01-02"))
	filePath := filepath.Join(filepath.Dir(l.filePath), filename)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	l.file = file
	l.filePath = filePath

	return nil
}

// CleanupOldLogs removes audit logs older than retention period
func CleanupOldLogs(retentionDays int) error {
	auditDir := filepath.Join(config.GetStateDir(), "audit")
	
	entries, err := os.ReadDir(auditDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(auditDir, entry.Name())
			os.Remove(path)
		}
	}

	return nil
}

// LogJobExecution logs a job execution event
func (l *Logger) LogJobExecution(jobID, jobType, status, command string, policyVersion int) error {
	return l.Log(EventJobExecuted, map[string]interface{}{
		"job_id":         jobID,
		"job_type":       jobType,
		"status":         status,
		"command":        command,
		"policy_version": policyVersion,
	})
}

// LogPolicyChange logs a policy change event
func (l *Logger) LogPolicyChange(oldVersion, newVersion int) error {
	return l.Log(EventPolicyChanged, map[string]interface{}{
		"old_version": oldVersion,
		"new_version": newVersion,
	})
}

// LogCertRotation logs a certificate rotation event
func (l *Logger) LogCertRotation(success bool, reason string) error {
	status := "success"
	if !success {
		status = "failed"
	}

	return l.Log(EventCertRotated, map[string]interface{}{
		"status": status,
		"reason": reason,
	})
}

// LogUpdate logs an update application event
func (l *Logger) LogUpdate(version string, success bool) error {
	status := "success"
	if !success {
		status = "failed"
	}

	return l.Log(EventUpdateApplied, map[string]interface{}{
		"version": version,
		"status":  status,
	})
}

// LogPolicyViolation logs a policy violation event
func (l *Logger) LogPolicyViolation(violationType, resource string, jobID string) error {
	return l.Log(EventPolicyViolation, map[string]interface{}{
		"violation_type": violationType,
		"resource":       resource,
		"job_id":         jobID,
	})
}
