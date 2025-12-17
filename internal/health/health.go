package health

import (
	"sync"
	"time"
)

// Status represents the health status
type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// Check represents an individual health check
type Check struct {
	Status      Status `json:"status"`
	Message     string `json:"message"`
	ExpiresInDays *int  `json:"expires_in_days,omitempty"`
}

// Report represents the overall health report
type Report struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]*Check `json:"checks"`
	Version   string            `json:"version"`
	AgentID   string            `json:"agent_id"`
}

// Checker performs health checks
type Checker struct {
	mu      sync.RWMutex
	checks  map[string]*Check
	version string
	agentID string
}

// NewChecker creates a new health checker
func NewChecker(version, agentID string) *Checker {
	return &Checker{
		checks:  make(map[string]*Check),
		version: version,
		agentID: agentID,
	}
}

// UpdateCheck updates a specific health check
func (c *Checker) UpdateCheck(name string, check *Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// GetReport generates a health report
func (c *Checker) GetReport() *Report {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Copy checks
	checks := make(map[string]*Check)
	overallHealthy := true
	
	for name, check := range c.checks {
		checks[name] = check
		if check.Status == StatusFail {
			overallHealthy = false
		}
	}

	status := "healthy"
	if !overallHealthy {
		status = "unhealthy"
	}

	return &Report{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
		Version:   c.version,
		AgentID:   c.agentID,
	}
}

// IsHealthy returns true if all checks pass
func (c *Checker) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, check := range c.checks {
		if check.Status == StatusFail {
			return false
		}
	}
	return true
}
