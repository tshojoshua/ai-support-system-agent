package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/sysinfo"
	"github.com/tshojoshua/jtnt-agent/internal/transport"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

const (
	heartbeatPath = "/api/v1/agent/heartbeat"
)

// sendHeartbeat sends a heartbeat to the hub
func (a *Agent) sendHeartbeat(ctx context.Context) error {
	// Collect system info
	sysInfo, err := a.sysinfo.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect system info: %w", err)
	}

	// Create heartbeat request
	req := api.HeartbeatRequest{
		AgentID:   a.config.AgentID,
		Timestamp: time.Now(),
		SysInfo:   *sysInfo,
	}

	// Send heartbeat
	respData, err := a.client.Post(ctx, heartbeatPath, req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	// Parse response
	var resp api.HeartbeatResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("failed to parse heartbeat response: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("heartbeat not acknowledged")
	}

	// Update heartbeat interval if changed
	if resp.NextHeartbeatSec > 0 && resp.NextHeartbeatSec != a.config.HeartbeatSec {
		a.logger.Info("heartbeat", map[string]interface{}{
			"message":             "heartbeat interval updated",
			"old_interval":        a.config.HeartbeatSec,
			"new_interval":        resp.NextHeartbeatSec,
		})
		a.config.HeartbeatSec = resp.NextHeartbeatSec
	}

	return nil
}

// heartbeatLoop continuously sends heartbeats
func (a *Agent) heartbeatLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(a.config.HeartbeatSec) * time.Second)
	defer ticker.Stop()

	a.logger.Info("heartbeat", map[string]interface{}{
		"message":  "heartbeat loop started",
		"interval": a.config.HeartbeatSec,
	})

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("heartbeat", map[string]interface{}{
				"message": "heartbeat loop stopped",
			})
			return

		case <-ticker.C:
			start := time.Now()
			
			if err := a.sendHeartbeat(a.ctx); err != nil {
				a.logger.Error("heartbeat", map[string]interface{}{
					"message": "heartbeat failed",
					"error":   err.Error(),
				})
			} else {
				a.logger.Debug("heartbeat", map[string]interface{}{
					"message":     "heartbeat sent successfully",
					"duration_ms": time.Since(start).Milliseconds(),
				})
			}
			
			// Update ticker if interval changed
			if ticker.Reset(time.Duration(a.config.HeartbeatSec) * time.Second) {
				// Ticker was reset successfully
			}
		}
	}
}
