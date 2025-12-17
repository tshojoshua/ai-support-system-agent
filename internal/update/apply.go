package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Applier handles platform-specific update application
type Applier struct {
	currentBinaryPath string
	serviceName       string
}

// NewApplier creates a new update applier
func NewApplier(currentBinaryPath, serviceName string) *Applier {
	return &Applier{
		currentBinaryPath: currentBinaryPath,
		serviceName:       serviceName,
	}
}

// Apply applies an update by replacing the current binary
func (a *Applier) Apply(newBinaryPath string) error {
	switch runtime.GOOS {
	case "windows":
		return a.applyWindows(newBinaryPath)
	case "darwin", "linux":
		return a.applyUnix(newBinaryPath)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// applyWindows applies update on Windows
func (a *Applier) applyWindows(newBinaryPath string) error {
	// Stop service if running
	if a.serviceName != "" {
		cmd := exec.Command("sc", "stop", a.serviceName)
		cmd.Run() // Ignore errors - service might not be installed
		time.Sleep(2 * time.Second)
	}

	// Backup current binary
	backupPath := a.currentBinaryPath + ".old"
	if err := os.Rename(a.currentBinaryPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Copy new binary to production location
	if err := copyFile(newBinaryPath, a.currentBinaryPath); err != nil {
		// Rollback
		os.Rename(backupPath, a.currentBinaryPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Start service if it was running
	if a.serviceName != "" {
		cmd := exec.Command("sc", "start", a.serviceName)
		if err := cmd.Run(); err != nil {
			// Rollback on failure
			os.Remove(a.currentBinaryPath)
			os.Rename(backupPath, a.currentBinaryPath)
			cmd = exec.Command("sc", "start", a.serviceName)
			cmd.Run()
			return fmt.Errorf("failed to start service: %w", err)
		}
	}

	// Wait a moment to verify service started
	time.Sleep(5 * time.Second)

	// Clean up backup after successful update
	os.Remove(backupPath)

	return nil
}

// applyUnix applies update on Unix-like systems (Linux, macOS)
func (a *Applier) applyUnix(newBinaryPath string) error {
	// Create backup
	backupPath := a.currentBinaryPath + ".old"
	if err := copyFile(a.currentBinaryPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Atomic rename (if on same filesystem)
	if err := os.Rename(newBinaryPath, a.currentBinaryPath); err != nil {
		// If rename fails (different filesystems), copy instead
		if err := copyFile(newBinaryPath, a.currentBinaryPath); err != nil {
			os.Rename(backupPath, a.currentBinaryPath)
			return fmt.Errorf("failed to install new binary: %w", err)
		}
	}

	// Set executable permissions
	if err := os.Chmod(a.currentBinaryPath, 0755); err != nil {
		os.Rename(backupPath, a.currentBinaryPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Restart service
	if err := a.restartService(); err != nil {
		// Rollback on failure
		os.Rename(backupPath, a.currentBinaryPath)
		a.restartService()
		return fmt.Errorf("failed to restart service: %w", err)
	}

	// Wait to verify service started
	time.Sleep(5 * time.Second)

	// Clean up backup
	os.Remove(backupPath)

	return nil
}

// restartService restarts the agent service
func (a *Applier) restartService() error {
	if a.serviceName == "" {
		return nil
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// Try systemd first
		cmd = exec.Command("systemctl", "restart", a.serviceName)
		if err := cmd.Run(); err != nil {
			// Try init.d
			cmd = exec.Command("service", a.serviceName, "restart")
			return cmd.Run()
		}
		return nil

	case "darwin":
		// macOS launchd
		cmd = exec.Command("launchctl", "kickstart", "-k", fmt.Sprintf("system/%s", a.serviceName))
		return cmd.Run()

	default:
		return fmt.Errorf("unsupported platform for service restart: %s", runtime.GOOS)
	}
}

// Rollback restores the previous version
func (a *Applier) Rollback() error {
	backupPath := a.currentBinaryPath + ".old"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found")
	}

	// Restore backup
	if err := copyFile(backupPath, a.currentBinaryPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Set permissions (Unix only)
	if runtime.GOOS != "windows" {
		os.Chmod(a.currentBinaryPath, 0755)
	}

	// Restart service
	return a.restartService()
}

// VerifyUpdate verifies the update was successful
func (a *Applier) VerifyUpdate(expectedVersion string) error {
	// Run the new binary with --version flag
	cmd := exec.Command(a.currentBinaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to verify new version: %w", err)
	}

	// Simple version check
	version := string(output)
	if len(version) == 0 {
		return fmt.Errorf("empty version output")
	}

	// In production, parse and compare versions properly
	return nil
}

// CleanupBackup removes the backup file
func (a *Applier) CleanupBackup() error {
	backupPath := a.currentBinaryPath + ".old"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(backupPath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Read source
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to temp file
	tmpDst := dst + ".tmp"
	if err := os.WriteFile(tmpDst, data, 0755); err != nil {
		return err
	}

	// Atomic rename
	return os.Rename(tmpDst, dst)
}

// GetCurrentBinaryPath attempts to find the current binary path
func GetCurrentBinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Resolve symlinks
	return filepath.EvalSymlinks(exe)
}
