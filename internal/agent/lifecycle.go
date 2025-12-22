package agent

import (
	"fmt"

	"github.com/tshojoshua/jtnt-agent/internal/config"
)

// Lifecycle management functions for the agent
// This module handles start, stop, reload operations

// Reload reloads the agent configuration
func (a *Agent) Reload() error {
	a.logger.Info("lifecycle", map[string]interface{}{
		"message": "reloading agent configuration",
	})

	// Load new configuration
	newConfig, err := config.Load(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Validate new configuration
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Update configuration
	a.config = newConfig

	a.logger.Info("lifecycle", map[string]interface{}{
		"message": "configuration reloaded successfully",
	})

	return nil
}

// Status returns the current agent status
func (a *Agent) Status() map[string]interface{} {
	return map[string]interface{}{
		"agent_id":      a.config.AgentID,
		"hub_url":       a.config.HubURL,
		"heartbeat_sec": a.config.HeartbeatSec,
		"poll_interval": a.config.PollIntervalSec,
	}
}
