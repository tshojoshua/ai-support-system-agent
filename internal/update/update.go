package update

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Info represents update information from hub
type Info struct {
	LatestVersion string    `json:"latest_version"`
	DownloadURL   string    `json:"download_url"`
	SignatureURL  string    `json:"signature_url"`
	SHA256        string    `json:"sha256"`
	ReleaseNotes  string    `json:"release_notes"`
	Critical      bool      `json:"critical"`
	PublishedAt   time.Time `json:"published_at"`
}

// Updater handles agent self-updates
type Updater struct {
	currentVersion string
	updateDir      string
	publicKey      ed25519.PublicKey
	httpClient     *http.Client
}

// NewUpdater creates a new updater
func NewUpdater(currentVersion, updateDir string, publicKey ed25519.PublicKey) *Updater {
	return &Updater{
		currentVersion: currentVersion,
		updateDir:      updateDir,
		publicKey:      publicKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// CheckForUpdate checks if an update is available
func (u *Updater) CheckForUpdate(ctx context.Context, hubURL string) (*Info, error) {
	url := fmt.Sprintf("%s/api/v1/agent/update/check", hubURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// No update available
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var info Info
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Replace placeholders in URLs
	info.DownloadURL = u.expandURL(info.DownloadURL)
	info.SignatureURL = u.expandURL(info.SignatureURL)

	return &info, nil
}

// expandURL replaces {os} and {arch} placeholders
func (u *Updater) expandURL(url string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	
	// Replace in simple way (could use strings.ReplaceAll in production)
	result := url
	result = replaceAll(result, "{os}", goos)
	result = replaceAll(result, "{arch}", goarch)
	return result
}

func replaceAll(s, old, new string) string {
	// Simple string replacement
	for {
		if len(s) < len(old) {
			return s
		}
		idx := -1
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == -1 {
			return s
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
}

// DownloadAndVerify downloads and verifies the update
func (u *Updater) DownloadAndVerify(ctx context.Context, info *Info) (string, error) {
	// Ensure update directory exists
	if err := os.MkdirAll(u.updateDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create update directory: %w", err)
	}

	// Download binary
	binaryPath := filepath.Join(u.updateDir, "jtnt-agentd.new")
	if err := u.downloadFile(ctx, info.DownloadURL, binaryPath); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	// Verify SHA256
	if err := u.verifySHA256(binaryPath, info.SHA256); err != nil {
		os.Remove(binaryPath)
		return "", fmt.Errorf("SHA256 verification failed: %w", err)
	}

	// Download signature
	sigPath := filepath.Join(u.updateDir, "jtnt-agentd.sig")
	if err := u.downloadFile(ctx, info.SignatureURL, sigPath); err != nil {
		os.Remove(binaryPath)
		return "", fmt.Errorf("failed to download signature: %w", err)
	}

	// Verify signature
	if err := u.verifySignature(binaryPath, sigPath); err != nil {
		os.Remove(binaryPath)
		os.Remove(sigPath)
		return "", fmt.Errorf("signature verification failed: %w", err)
	}

	// Clean up signature file
	os.Remove(sigPath)

	return binaryPath, nil
}

// downloadFile downloads a file from URL
func (u *Updater) downloadFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Write to temp file first
	tempPath := destPath + ".tmp"
	f, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(tempPath)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}

	// Atomic rename
	return os.Rename(tempPath, destPath)
}

// verifySHA256 verifies the SHA256 checksum
func (u *Updater) verifySHA256(filePath, expectedHash string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(h.Sum(nil))

	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

// verifySignature verifies the Ed25519 signature
func (u *Updater) verifySignature(filePath, sigPath string) error {
	// Read binary
	binary, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	// Read signature
	signature, err := os.ReadFile(sigPath)
	if err != nil {
		return fmt.Errorf("failed to read signature: %w", err)
	}

	// Verify
	if !ed25519.Verify(u.publicKey, binary, signature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// NeedsUpdate returns true if the info represents a newer version
func (u *Updater) NeedsUpdate(info *Info) bool {
	if info == nil {
		return false
	}

	// Simple version comparison (in production, use semver)
	return info.LatestVersion != u.currentVersion
}

// CleanupOldUpdates removes old update files
func (u *Updater) CleanupOldUpdates() error {
	if _, err := os.Stat(u.updateDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(u.updateDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		path := filepath.Join(u.updateDir, entry.Name())
		os.Remove(path)
	}

	return nil
}
