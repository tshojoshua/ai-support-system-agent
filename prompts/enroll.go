package enroll

import (
	"context"
	"fmt"
	"time"

	"github.com/jtnt/agent/internal/config"
	"github.com/jtnt/agent/internal/sysinfo"
	"github.com/jtnt/agent/internal/transport"
	"github.com/jtnt/agent/pkg/api"
)

func Enroll(hubURL, token string) error {
	// Collect system information
	sysInfo, err := sysinfo.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	// Prepare enrollment request
	req := api.EnrollRequest{
		Token:        token,
		Hostname:     sysInfo.Hostname,
		OS:           sysInfo.OS,
		OSVersion:    sysInfo.OSVersion,
		Arch:         sysInfo.Arch,
		AgentVersion: config.Version,
		Capabilities: []string{
			"system_info",
			"remote_exec",
			"file_transfer",
			"monitoring",
		},
		SystemInfo: map[string]interface{}{
			"cpu_count":    sysInfo.CPUCount,
			"mem_total":    sysInfo.MemTotal,
			"disk_total":   sysInfo.DiskTotal,
			"ip_addresses": sysInfo.IPAddresses,
		},
	}

	// Create HTTP client
	client := transport.NewClient(hubURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Send enrollment request
	var resp api.EnrollResponse
	if err := client.Post(ctx, "/api/v1/agents/enroll", req, &resp); err != nil {
		return fmt.Errorf("enrollment failed: %w", err)
	}

	// Save configuration
	cfg := &config.Config{
		AgentID:              resp.AgentID,
		HubURL:               resp.HubBaseURL,
		AgentToken:           resp.AgentToken,
		PollIntervalSec:      resp.PollIntervalSec,
		HeartbeatIntervalSec: resp.HeartbeatIntervalSec,
		TenantID:             resp.TenantID,
		SiteID:               resp.SiteID,
		EnrolledAt:           time.Now(),
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Agent enrolled successfully!\n")
	fmt.Printf("Agent ID: %s\n", resp.AgentID)
	fmt.Printf("Tenant ID: %s\n", resp.TenantID)
	if resp.SiteID != "" {
		fmt.Printf("Site ID: %s\n", resp.SiteID)
	}
	fmt.Printf("\nConfiguration saved to: %s\n", config.GetConfigPath())

	return nil
}
