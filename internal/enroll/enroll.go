package enroll

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
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
	enrollPath = "/api/v1/agent/enroll"
	agentVersion = "1.0.0"
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
		Token:     token,
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Version:   agentVersion,
		PublicKey: keypair.PublicKeyBase64(),
	}

	// Send enrollment request
	resp, err := e.sendEnrollRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("enrollment request failed: %w", err)
	}

	// Validate certificate chain
	if err := e.validateCertificates(resp); err != nil {
		return nil, fmt.Errorf("certificate validation failed: %w", err)
	}

	// Save certificates
	if err := e.saveCertificates(resp); err != nil {
		return nil, fmt.Errorf("failed to save certificates: %w", err)
	}

	// Create configuration
	cfg := &config.Config{
		AgentID:         resp.AgentID,
		HubURL:          resp.HubBaseURL,
		PollIntervalSec: resp.PollIntervalSec,
		HeartbeatSec:    resp.HeartbeatSec,
		CertPath:        config.GetCertPath(),
		KeyPath:         config.GetKeyPath(),
		CABundlePath:    config.GetCABundlePath(),
		PolicyVersion:   resp.Policy.Version,
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

func (e *Enroller) validateCertificates(resp *api.EnrollResponse) error {
	// Parse client certificate
	certPEM := []byte(resp.ClientCertPEM)
	keyPEM := []byte(resp.ClientKeyPEM)
	
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("invalid client certificate: %w", err)
	}

	// Parse CA bundle
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM([]byte(resp.CABundlePEM)) {
		return fmt.Errorf("failed to parse CA bundle")
	}

	// Parse and verify client certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Verify certificate chain
	opts := x509.VerifyOptions{
		Roots:     caPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if _, err := x509Cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

func (e *Enroller) saveCertificates(resp *api.EnrollResponse) error {
	certsDir := config.GetCertsDir()
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Save client certificate
	if err := e.store.Save("client.crt", []byte(resp.ClientCertPEM)); err != nil {
		return fmt.Errorf("failed to save client cert: %w", err)
	}

	// Save client key
	if err := e.store.Save("client.key", []byte(resp.ClientKeyPEM)); err != nil {
		return fmt.Errorf("failed to save client key: %w", err)
	}

	// Save CA bundle
	if err := e.store.Save("ca-bundle.crt", []byte(resp.CABundlePEM)); err != nil {
		return fmt.Errorf("failed to save CA bundle: %w", err)
	}

	return nil
}
