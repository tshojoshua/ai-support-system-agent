package agent

import (
	"context"
	"log"
	"time"

	"github.com/jtnt/agent/internal/sysinfo"
	"github.com/jtnt/agent/internal/transport"
	"github.com/jtnt/agent/pkg/api"
)

func (a *Agent) heartbeatLoop() {
	defer a.wg.Done()

	interval := time.Duration(a.config.HeartbeatIntervalSec) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting heartbeat loop (interval: %v)", interval)

	// Send initial heartbeat immediately
	a.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			a.sendHeartbeat()
		case <-a.ctx.Done():
			log.Println("Heartbeat loop stopped")
			return
		}
	}
}

func (a *Agent) sendHeartbeat() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Collect current system info
	sysInfo, err := sysinfo.Collect()
	if err != nil {
		log.Printf("Failed to collect system info: %v", err)
		return
	}

	// Prepare heartbeat request
	req := api.HeartbeatRequest{
		AgentID:   a.config.AgentID,
		Timestamp: time.Now(),
		Sysinfo: map[string]interface{}{
			"hostname":     sysInfo.Hostname,
			"os":           sysInfo.OS,
			"os_version":   sysInfo.OSVersion,
			"arch":         sysInfo.Arch,
			"uptime":       sysInfo.Uptime,
			"cpu_count":    sysInfo.CPUCount,
			"cpu_usage":    sysInfo.CPUUsage,
			"mem_total":    sysInfo.MemTotal,
			"mem_used":     sysInfo.MemUsed,
			"mem_percent":  sysInfo.MemPercent,
			"disk_total":   sysInfo.DiskTotal,
			"disk_used":    sysInfo.DiskUsed,
			"disk_percent": sysInfo.DiskPercent,
			"ip_addresses": sysInfo.IPAddresses,
		},
	}

	// Send heartbeat with retry
	retryConfig := transport.DefaultRetryConfig
	retryConfig.MaxAttempts = 3 // Limit retries for individual heartbeat

	err = transport.WithRetry(ctx, retryConfig, func() error {
		var resp api.HeartbeatResponse
		if err := a.client.Post(ctx, "/api/v1/agents/heartbeat", req, &resp); err != nil {
			return err
		}

		// Update interval if hub requests it
		if resp.NextHeartbeatSec > 0 && resp.NextHeartbeatSec != a.config.HeartbeatIntervalSec {
			log.Printf("Heartbeat interval updated: %ds -> %ds",
				a.config.HeartbeatIntervalSec, resp.NextHeartbeatSec)
			a.config.HeartbeatIntervalSec = resp.NextHeartbeatSec
			a.config.Save()
		}

		return nil
	})

	if err != nil {
		log.Printf("Heartbeat failed: %v", err)
	} else {
		log.Printf("Heartbeat sent successfully (CPU: %.1f%%, Mem: %.1f%%)",
			sysInfo.CPUUsage, sysInfo.MemPercent)
	}
}
