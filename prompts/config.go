package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const Version = "1.0.0"

type Config struct {
	AgentID              string    `json:"agent_id"`
	DeviceID             string    `json:"device_id,omitempty"`
	HubURL               string    `json:"hub_url"`
	AgentToken           string    `json:"agent_token"`
	PollIntervalSec      int       `json:"poll_interval_sec"`
	HeartbeatIntervalSec int       `json:"heartbeat_interval_sec"`
	TenantID             string    `json:"tenant_id,omitempty"`
	SiteID               string    `json:"site_id,omitempty"`
	EnrolledAt           time.Time `json:"enrolled_at"`
}

func Load() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(GetStateDir(), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func Exists() bool {
	_, err := os.Stat(GetConfigPath())
	return err == nil
}
