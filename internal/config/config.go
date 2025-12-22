package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds agent configuration
type Config struct {
	AgentID         string `json:"agent_id"`
	AgentToken      string `json:"agent_token"`
	HubURL          string `json:"hub_url"`
	PollIntervalSec int    `json:"poll_interval_sec"`
	HeartbeatSec    int    `json:"heartbeat_sec"`
}

// Load reads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes configuration to file
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(GetConfigDir(), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	if c.HubURL == "" {
		return fmt.Errorf("hub_url is required")
	}
	if c.AgentToken == "" {
		return fmt.Errorf("agent_token is required")
	}
	return nil
}
