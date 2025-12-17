package certman

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	certRenewalThreshold = 30 * 24 * time.Hour // 30 days
	certBackupSuffix     = ".backup"
	certNewSuffix        = ".new"
	certBackupRetention  = 7 * 24 * time.Hour // 7 days
)

// Manager handles certificate rotation and renewal
type Manager struct {
	certPath     string
	keyPath      string
	caBundlePath string
	privateKey   ed25519.PrivateKey
}

// NewManager creates a new certificate manager
func NewManager(certPath, keyPath, caBundlePath string, privateKey ed25519.PrivateKey) *Manager {
	return &Manager{
		certPath:     certPath,
		keyPath:      keyPath,
		caBundlePath: caBundlePath,
		privateKey:   privateKey,
	}
}

// CheckExpiration checks if certificate needs renewal
func (m *Manager) CheckExpiration() (needsRenewal bool, daysUntilExpiry int, err error) {
	certPEM, err := os.ReadFile(m.certPath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false, 0, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse certificate: %w", err)
	}

	timeUntilExpiry := time.Until(cert.NotAfter)
	daysUntilExpiry = int(timeUntilExpiry.Hours() / 24)

	needsRenewal = timeUntilExpiry <= certRenewalThreshold

	return needsRenewal, daysUntilExpiry, nil
}

// GetCertificateSerial returns the current certificate serial number
func (m *Manager) GetCertificateSerial() (string, error) {
	certPEM, err := os.ReadFile(m.certPath)
	if err != nil {
		return "", fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert.SerialNumber.String(), nil
}

// GenerateCSR generates a Certificate Signing Request
func (m *Manager) GenerateCSR(agentID string) ([]byte, error) {
	// Generate new key pair for CSR
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create CSR template
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   agentID,
			Organization: []string{"JTNT"},
		},
		SignatureAlgorithm: x509.PureEd25519,
	}

	// Create CSR
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %w", err)
	}

	// Encode to PEM
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	return csrPEM, nil
}

// InstallNewCertificate installs a new certificate atomically
func (m *Manager) InstallNewCertificate(certPEM, caBundlePEM []byte) error {
	// Validate certificate before installing
	if err := m.validateCertificate(certPEM, caBundlePEM); err != nil {
		return fmt.Errorf("certificate validation failed: %w", err)
	}

	// Backup current certificate
	if err := m.backupCertificate(); err != nil {
		return fmt.Errorf("failed to backup certificate: %w", err)
	}

	// Write new certificate to .new file
	newCertPath := m.certPath + certNewSuffix
	if err := os.WriteFile(newCertPath, certPEM, 0600); err != nil {
		return fmt.Errorf("failed to write new certificate: %w", err)
	}

	// Atomic rename
	if err := os.Rename(newCertPath, m.certPath); err != nil {
		return fmt.Errorf("failed to install certificate: %w", err)
	}

	// Update CA bundle if provided
	if len(caBundlePEM) > 0 {
		newCAPath := m.caBundlePath + certNewSuffix
		if err := os.WriteFile(newCAPath, caBundlePEM, 0644); err != nil {
			return fmt.Errorf("failed to write CA bundle: %w", err)
		}

		if err := os.Rename(newCAPath, m.caBundlePath); err != nil {
			return fmt.Errorf("failed to install CA bundle: %w", err)
		}
	}

	return nil
}

// validateCertificate validates the new certificate
func (m *Manager) validateCertificate(certPEM, caBundlePEM []byte) error {
	// Parse certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check expiration
	now := time.Now()
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate is expired")
	}

	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}

	// Parse CA bundle
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caBundlePEM) {
		return fmt.Errorf("failed to parse CA bundle")
	}

	// Verify certificate chains to CA
	opts := x509.VerifyOptions{
		Roots:     caPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// backupCertificate creates a backup of the current certificate
func (m *Manager) backupCertificate() error {
	backupPath := m.certPath + certBackupSuffix

	// Read current certificate
	certData, err := os.ReadFile(m.certPath)
	if err != nil {
		return fmt.Errorf("failed to read current certificate: %w", err)
	}

	// Write backup
	if err := os.WriteFile(backupPath, certData, 0600); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// Rollback restores the previous certificate from backup
func (m *Manager) Rollback() error {
	backupPath := m.certPath + certBackupSuffix

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup certificate found")
	}

	// Restore backup
	if err := os.Rename(backupPath, m.certPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// CleanupOldBackups removes old backup files
func (m *Manager) CleanupOldBackups() error {
	backupPath := m.certPath + certBackupSuffix

	info, err := os.Stat(backupPath)
	if os.IsNotExist(err) {
		return nil // No backup to clean
	}
	if err != nil {
		return err
	}

	// Check backup age
	if time.Since(info.ModTime()) > certBackupRetention {
		return os.Remove(backupPath)
	}

	return nil
}

// GetExpirationTime returns the certificate expiration time
func (m *Manager) GetExpirationTime() (time.Time, error) {
	certPEM, err := os.ReadFile(m.certPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert.NotAfter, nil
}

// ShouldCheckRenewal returns true if it's time to check for renewal
func ShouldCheckRenewal(lastCheck time.Time) bool {
	// Check daily at 03:00
	now := time.Now()
	
	// If never checked, check now
	if lastCheck.IsZero() {
		return true
	}

	// If last check was yesterday or earlier, and it's past 03:00, check now
	if now.Sub(lastCheck) >= 24*time.Hour && now.Hour() >= 3 {
		return true
	}

	return false
}

// GetCertificatePath returns the path to the certificate file
func (m *Manager) GetCertificatePath() string {
	return m.certPath
}

// GetKeyPath returns the path to the private key file
func (m *Manager) GetKeyPath() string {
	return m.keyPath
}

// GetCABundlePath returns the path to the CA bundle file
func (m *Manager) GetCABundlePath() string {
	return m.caBundlePath
}
