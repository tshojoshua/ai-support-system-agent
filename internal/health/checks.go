package health

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/tshojoshua/jtnt-agent/internal/config"
)

const (
	certWarningDays      = 30
	policyWarningDays    = 7
	heartbeatMaxAge      = 5 * time.Minute
	diskUsageWarningPct  = 90.0
)

// CheckEnrolled checks if the agent is properly enrolled
func CheckEnrolled(cfg *config.Config) *Check {
	if cfg == nil {
		return &Check{
			Status:  StatusFail,
			Message: "configuration not loaded",
		}
	}

	if cfg.AgentID == "" {
		return &Check{
			Status:  StatusFail,
			Message: "agent not enrolled - missing agent ID",
		}
	}

	// Check if certificates exist
	if _, err := os.Stat(cfg.CertPath); os.IsNotExist(err) {
		return &Check{
			Status:  StatusFail,
			Message: "agent not enrolled - missing client certificate",
		}
	}

	if _, err := os.Stat(cfg.KeyPath); os.IsNotExist(err) {
		return &Check{
			Status:  StatusFail,
			Message: "agent not enrolled - missing private key",
		}
	}

	return &Check{
		Status:  StatusPass,
		Message: "agent enrolled",
	}
}

// CheckCertificates checks certificate validity and expiration
func CheckCertificates(certPath string) *Check {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return &Check{
			Status:  StatusFail,
			Message: fmt.Sprintf("failed to read certificate: %v", err),
		}
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return &Check{
			Status:  StatusFail,
			Message: "failed to parse certificate PEM",
		}
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return &Check{
			Status:  StatusFail,
			Message: fmt.Sprintf("failed to parse certificate: %v", err),
		}
	}

	now := time.Now()

	// Check if expired
	if now.After(cert.NotAfter) {
		return &Check{
			Status:  StatusFail,
			Message: fmt.Sprintf("certificate expired on %s", cert.NotAfter.Format("2006-01-02")),
		}
	}

	// Check if not yet valid
	if now.Before(cert.NotBefore) {
		return &Check{
			Status:  StatusFail,
			Message: fmt.Sprintf("certificate not yet valid (valid from %s)", cert.NotBefore.Format("2006-01-02")),
		}
	}

	// Calculate days until expiration
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

	if daysUntilExpiry <= certWarningDays {
		return &Check{
			Status:        StatusWarn,
			Message:       fmt.Sprintf("certificate expires in %d days", daysUntilExpiry),
			ExpiresInDays: &daysUntilExpiry,
		}
	}

	return &Check{
		Status:        StatusPass,
		Message:       fmt.Sprintf("certificate valid until %s", cert.NotAfter.Format("2006-01-02")),
		ExpiresInDays: &daysUntilExpiry,
	}
}

// CheckHubConnection checks hub connectivity based on last heartbeat
func CheckHubConnection(lastHeartbeat time.Time) *Check {
	if lastHeartbeat.IsZero() {
		return &Check{
			Status:  StatusFail,
			Message: "no heartbeat sent yet",
		}
	}

	timeSinceHeartbeat := time.Since(lastHeartbeat)

	if timeSinceHeartbeat > heartbeatMaxAge {
		return &Check{
			Status:  StatusFail,
			Message: fmt.Sprintf("no heartbeat for %.0f seconds", timeSinceHeartbeat.Seconds()),
		}
	}

	return &Check{
		Status:  StatusPass,
		Message: fmt.Sprintf("last heartbeat %.0f seconds ago", timeSinceHeartbeat.Seconds()),
	}
}

// CheckPolicy checks policy validity and expiration
func CheckPolicy(policyExpiresAt time.Time) *Check {
	if policyExpiresAt.IsZero() {
		return &Check{
			Status:  StatusWarn,
			Message: "policy expiration not set",
		}
	}

	now := time.Now()

	if now.After(policyExpiresAt) {
		return &Check{
			Status:  StatusFail,
			Message: fmt.Sprintf("policy expired on %s", policyExpiresAt.Format("2006-01-02")),
		}
	}

	daysUntilExpiry := int(time.Until(policyExpiresAt).Hours() / 24)

	if daysUntilExpiry <= policyWarningDays {
		return &Check{
			Status:  StatusWarn,
			Message: fmt.Sprintf("policy expires in %d days", daysUntilExpiry),
		}
	}

	return &Check{
		Status:  StatusPass,
		Message: fmt.Sprintf("policy valid until %s", policyExpiresAt.Format("2006-01-02")),
	}
}

// CheckDiskSpace checks disk space for state directory
func CheckDiskSpace() *Check {
	stateDir := config.GetStateDir()
	
	usage, err := disk.Usage(filepath.Dir(stateDir))
	if err != nil {
		return &Check{
			Status:  StatusWarn,
			Message: fmt.Sprintf("failed to check disk space: %v", err),
		}
	}

	usedPercent := usage.UsedPercent

	if usedPercent >= diskUsageWarningPct {
		return &Check{
			Status:  StatusWarn,
			Message: fmt.Sprintf("disk space %.1f%% used (warning threshold: %.0f%%)", usedPercent, diskUsageWarningPct),
		}
	}

	return &Check{
		Status:  StatusPass,
		Message: fmt.Sprintf("disk space %.1f%% used", usedPercent),
	}
}

// CheckLastJob checks the status of the last job execution
func CheckLastJob(lastJobStatus string, lastJobTime time.Time) *Check {
	if lastJobTime.IsZero() {
		return &Check{
			Status:  StatusPass,
			Message: "no jobs executed yet",
		}
	}

	timeSinceJob := time.Since(lastJobTime)

	if lastJobStatus == "failed" {
		return &Check{
			Status:  StatusWarn,
			Message: fmt.Sprintf("last job failed %.0f minutes ago", timeSinceJob.Minutes()),
		}
	}

	return &Check{
		Status:  StatusPass,
		Message: fmt.Sprintf("last job completed %.0f minutes ago", timeSinceJob.Minutes()),
	}
}
