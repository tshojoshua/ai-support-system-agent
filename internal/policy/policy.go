package policy

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Policy represents agent execution policy
type Policy struct {
	Version      int          `json:"version"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Signature    string       `json:"signature"` // Ed25519 signature of policy JSON
	Capabilities Capabilities `json:"capabilities"`
}

// Capabilities defines what operations the agent can perform
type Capabilities struct {
	Exec   *ExecCapability   `json:"exec,omitempty"`
	Script *ScriptCapability `json:"script,omitempty"`
	File   *FileCapability   `json:"file,omitempty"`
}

// ExecCapability controls binary execution
type ExecCapability struct {
	Enabled            bool     `json:"enabled"`
	AllowedBinaries    []string `json:"allowed_binaries"`
	AllowedPaths       []string `json:"allowed_paths"` // Glob patterns
	MaxExecutionSec    int      `json:"max_execution_sec"`
	BlockNetworkAccess bool     `json:"block_network_access"`
}

// ScriptCapability controls script execution
type ScriptCapability struct {
	Enabled             bool     `json:"enabled"`
	AllowedInterpreters []string `json:"allowed_interpreters"`
	RequireSignature    bool     `json:"require_signature"`
	MaxScriptSizeBytes  int      `json:"max_script_size_bytes"`
	MaxExecutionSec     int      `json:"max_execution_sec"`
}

// FileCapability controls file operations
type FileCapability struct {
	ReadPaths       []string `json:"read_paths"`  // Glob patterns
	WritePaths      []string `json:"write_paths"` // Glob patterns
	MaxFileSizeBytes int64   `json:"max_file_size_bytes"`
}

// Load parses policy from JSON
func Load(data []byte) (*Policy, error) {
	var p Policy
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy: %w", err)
	}
	return &p, nil
}

// Validate checks if policy is valid and not expired
func (p *Policy) Validate() error {
	if p.Version < 1 {
		return fmt.Errorf("invalid policy version: %d", p.Version)
	}

	if time.Now().After(p.ExpiresAt) {
		return fmt.Errorf("policy expired at %s", p.ExpiresAt)
	}

	return nil
}

// VerifySignature verifies the policy signature
func (p *Policy) VerifySignature(publicKey ed25519.PublicKey) error {
	// Decode signature
	sig, err := base64.StdEncoding.DecodeString(p.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Create canonical JSON without signature for verification
	policyCopy := *p
	policyCopy.Signature = ""
	
	canonical, err := json.Marshal(policyCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	// Verify signature
	if !ed25519.Verify(publicKey, canonical, sig) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// DefaultPolicy returns a secure default policy
func DefaultPolicy() *Policy {
	return &Policy{
		Version:   1,
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		Capabilities: Capabilities{
			Exec: &ExecCapability{
				Enabled: true,
				AllowedBinaries: []string{
					"ipconfig", "whoami", "systeminfo", "hostname",
					"uname", "df", "ip", "ifconfig", "netstat",
					"system_profiler", "scutil",
				},
				AllowedPaths: []string{
					"C:\\Windows\\System32\\*",
					"/usr/bin/*",
					"/bin/*",
					"/usr/sbin/*",
					"/sbin/*",
				},
				MaxExecutionSec:    300,
				BlockNetworkAccess: false,
			},
			Script: &ScriptCapability{
				Enabled:             true,
				AllowedInterpreters: []string{"powershell", "bash", "sh"},
				RequireSignature:    true,
				MaxScriptSizeBytes:  1048576, // 1MB
				MaxExecutionSec:     600,
			},
			File: &FileCapability{
				ReadPaths: []string{
					"C:\\Logs\\*",
					"C:\\ProgramData\\JTNT\\*",
					"/var/log/*",
					"/tmp/jtnt/*",
					"/Library/Logs/*",
				},
				WritePaths: []string{
					"C:\\Temp\\JTNT\\*",
					"/tmp/jtnt/*",
					"/var/tmp/jtnt/*",
				},
				MaxFileSizeBytes: 104857600, // 100MB
			},
		},
	}
}

// ToJSON converts policy to JSON
func (p *Policy) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}
