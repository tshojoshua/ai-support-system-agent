package enroll

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/config"
	"github.com/tshojoshua/jtnt-agent/internal/store"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

const (
	enrollPath    = "/api/v1/agents/enroll"
	agentVersion  = "1.0.0"
	enrollTimeout = 30 * time.Second
)

// Enroller handles agent enrollment
type Enroller struct {
	hubURL string
	store  store.Store
}

// NewEnroller creates a new enroller
func NewEnroller(hubURL string, store store.Store) *Enroller {
	return &Enroller{
		hubURL: hubURL,
		store:  store,
	}
}

// Enroll performs the enrollment process
func (e *Enroller) Enroll(ctx context.Context, token string) (*config.Config, error) {
	// Generate Ed25519 keypair
	keypair, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Create enrollment request
	req := api.EnrollRequest{
		Token:        token,
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		Version:      agentVersion,
		AgentVersion: agentVersion,
		Capabilities: []string{"ping", "execute", "shell", "file_transfer"},
		PublicKey:    keypair.PublicKeyBase64(),
	}

	// Send enrollment request
	resp, err := e.sendEnrollRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("enrollment request failed: %w", err)
	}

	// Save agent token
	if err := e.saveAgentToken(resp.AgentToken); err != nil {
		return nil, fmt.Errorf("failed to save agent token: %w", err)
	}

	// Create configuration
	cfg := &config.Config{
		AgentID:         resp.AgentID,
		AgentToken:      resp.AgentToken,
		HubURL:          resp.HubBaseURL,
		PollIntervalSec: resp.PollIntervalSec,
		HeartbeatSec:    resp.HeartbeatSec,
	}

	// Save configuration
	if err := cfg.Save(config.GetConfigPath()); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return cfg, nil
}

func (e *Enroller) sendEnrollRequest(ctx context.Context, req api.EnrollRequest) (*api.EnrollResponse, error) {
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", e.hubURL+enrollPath, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: enrollTimeout,
	}

	// Send request
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("enrollment failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("enrollment failed with status %d", httpResp.StatusCode)
	}

	// Parse response
	var resp api.EnrollResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

func (e *Enroller) saveAgentToken(token string) error {
	// Save agent token to secure store
	if err := e.store.Save("agent.token", []byte(token)); err != nil {
		return fmt.Errorf("failed to save agent token: %w", err)
	}
	return nil
}
