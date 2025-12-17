package certman

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// RenewalClient defines the interface for certificate renewal API calls
type RenewalClient interface {
	RenewCertificate(ctx context.Context, req *RenewalRequest) (*RenewalResponse, error)
}

// RenewalRequest represents a certificate renewal request
type RenewalRequest struct {
	AgentID           string `json:"agent_id"`
	CurrentCertSerial string `json:"current_cert_serial"`
	CSR               string `json:"csr"` // base64-encoded CSR
}

// RenewalResponse represents a certificate renewal response
type RenewalResponse struct {
	ClientCertPEM string    `json:"client_cert_pem"`
	CABundlePEM   string    `json:"ca_bundle_pem"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// Renewer handles automatic certificate renewal
type Renewer struct {
	manager       *Manager
	client        RenewalClient
	agentID       string
	lastCheckTime time.Time
}

// NewRenewer creates a new certificate renewer
func NewRenewer(manager *Manager, client RenewalClient, agentID string) *Renewer {
	return &Renewer{
		manager: manager,
		client:  client,
		agentID: agentID,
	}
}

// CheckAndRenew checks if renewal is needed and performs it
func (r *Renewer) CheckAndRenew(ctx context.Context) error {
	// Check if certificate needs renewal
	needsRenewal, daysUntilExpiry, err := r.manager.CheckExpiration()
	if err != nil {
		return fmt.Errorf("failed to check expiration: %w", err)
	}

	if !needsRenewal {
		return nil
	}

	// Certificate needs renewal
	return r.Renew(ctx, fmt.Sprintf("certificate expires in %d days", daysUntilExpiry))
}

// Renew performs certificate renewal
func (r *Renewer) Renew(ctx context.Context, reason string) error {
	// Get current certificate serial
	serial, err := r.manager.GetCertificateSerial()
	if err != nil {
		return fmt.Errorf("failed to get certificate serial: %w", err)
	}

	// Generate CSR
	csrPEM, err := r.manager.GenerateCSR(r.agentID)
	if err != nil {
		return fmt.Errorf("failed to generate CSR: %w", err)
	}

	// Encode CSR to base64
	csrBase64 := base64.StdEncoding.EncodeToString(csrPEM)

	// Request renewal from hub
	req := &RenewalRequest{
		AgentID:           r.agentID,
		CurrentCertSerial: serial,
		CSR:               csrBase64,
	}

	resp, err := r.client.RenewCertificate(ctx, req)
	if err != nil {
		return fmt.Errorf("renewal request failed: %w", err)
	}

	// Install new certificate
	certPEM := []byte(resp.ClientCertPEM)
	caBundlePEM := []byte(resp.CABundlePEM)

	if err := r.manager.InstallNewCertificate(certPEM, caBundlePEM); err != nil {
		return fmt.Errorf("failed to install new certificate: %w", err)
	}

	r.lastCheckTime = time.Now()

	return nil
}

// ShouldCheck returns true if it's time to check for renewal
func (r *Renewer) ShouldCheck() bool {
	return ShouldCheckRenewal(r.lastCheckTime)
}

// GetLastCheckTime returns the last renewal check time
func (r *Renewer) GetLastCheckTime() time.Time {
	return r.lastCheckTime
}

// mockRenewalClient implements RenewalClient for testing
type mockRenewalClient struct {
	RenewFunc func(ctx context.Context, req *RenewalRequest) (*RenewalResponse, error)
}

func (m *mockRenewalClient) RenewCertificate(ctx context.Context, req *RenewalRequest) (*RenewalResponse, error) {
	if m.RenewFunc != nil {
		return m.RenewFunc(ctx, req)
	}
	return nil, fmt.Errorf("not implemented")
}
