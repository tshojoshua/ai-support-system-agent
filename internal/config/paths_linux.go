// +build !windows,!darwin

package config

import "path/filepath"

const (
	// StateDir is the directory for agent state
	StateDir = "/var/lib/jtnt-agent"
	
	// ConfigDir is the directory for agent configuration
	ConfigDir = "/etc/jtnt-agent"
	
	// BinaryPath is the path to the agent daemon
	BinaryPath = "/usr/local/bin/jtnt-agentd"
)

// GetStateDir returns the OS-specific state directory
func GetStateDir() string {
	return StateDir
}

// GetConfigDir returns the OS-specific config directory
func GetConfigDir() string {
	return ConfigDir
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() string {
	return filepath.Join(ConfigDir, "config.json")
}

// GetCertsDir returns the directory for certificates
func GetCertsDir() string {
	return filepath.Join(StateDir, "certs")
}

// GetCertPath returns the path to the client certificate
func GetCertPath() string {
	return filepath.Join(GetCertsDir(), "client.crt")
}

// GetKeyPath returns the path to the client key
func GetKeyPath() string {
	return filepath.Join(GetCertsDir(), "client.key")
}

// GetCABundlePath returns the path to the CA bundle
func GetCABundlePath() string {
	return filepath.Join(GetCertsDir(), "ca-bundle.crt")
}

// GetBinaryPath returns the path to the agent binary
func GetBinaryPath() string {
	return BinaryPath
}
