package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				AgentID:      "test-agent-id",
				HubURL:       "https://hub.example.com",
				CertPath:     "/path/to/cert",
				KeyPath:      "/path/to/key",
				CABundlePath: "/path/to/ca",
			},
			wantErr: false,
		},
		{
			name: "missing agent_id",
			config: Config{
				HubURL:       "https://hub.example.com",
				CertPath:     "/path/to/cert",
				KeyPath:      "/path/to/key",
				CABundlePath: "/path/to/ca",
			},
			wantErr: true,
		},
		{
			name: "missing hub_url",
			config: Config{
				AgentID:      "test-agent-id",
				CertPath:     "/path/to/cert",
				KeyPath:      "/path/to/key",
				CABundlePath: "/path/to/ca",
			},
			wantErr: true,
		},
		{
			name: "missing cert paths",
			config: Config{
				AgentID: "test-agent-id",
				HubURL:  "https://hub.example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "jtnt-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create test config
	originalCfg := Config{
		AgentID:         "test-agent-123",
		HubURL:          "https://hub.test.com",
		PollIntervalSec: 300,
		HeartbeatSec:    60,
		CertPath:        "/test/cert.crt",
		KeyPath:         "/test/key.key",
		CABundlePath:    "/test/ca.crt",
		PolicyVersion:   1,
	}

	// Save config
	if err := originalCfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load config
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Compare
	if loadedCfg.AgentID != originalCfg.AgentID {
		t.Errorf("AgentID mismatch: got %v, want %v", loadedCfg.AgentID, originalCfg.AgentID)
	}
	if loadedCfg.HubURL != originalCfg.HubURL {
		t.Errorf("HubURL mismatch: got %v, want %v", loadedCfg.HubURL, originalCfg.HubURL)
	}
	if loadedCfg.HeartbeatSec != originalCfg.HeartbeatSec {
		t.Errorf("HeartbeatSec mismatch: got %v, want %v", loadedCfg.HeartbeatSec, originalCfg.HeartbeatSec)
	}
}

func TestLoadNonexistentConfig(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error when loading nonexistent config")
	}
}
