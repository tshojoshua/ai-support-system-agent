package api

import "time"

// EnrollRequest is sent by agent during initial enrollment
type EnrollRequest struct {
	Token     string `json:"token"`
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Version   string `json:"version"`
	PublicKey string `json:"public_key"` // base64-encoded Ed25519 public key
}

// EnrollResponse is returned by hub after successful enrollment
type EnrollResponse struct {
	AgentID          string         `json:"agent_id"`
	HubBaseURL       string         `json:"hub_base_url"`
	PollIntervalSec  int            `json:"poll_interval_sec"`
	HeartbeatSec     int            `json:"heartbeat_every_sec"`
	ClientCertPEM    string         `json:"client_cert_pem"`
	ClientKeyPEM     string         `json:"client_key_pem"`
	CABundlePEM      string         `json:"ca_bundle_pem"`
	Policy           EnrollmentPolicy `json:"policy"`
}

// EnrollmentPolicy represents agent capabilities and restrictions
type EnrollmentPolicy struct {
	Version      int                    `json:"version"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// HeartbeatRequest is sent periodically by agent
type HeartbeatRequest struct {
	AgentID   string     `json:"agent_id"`
	Timestamp time.Time  `json:"timestamp"`
	SysInfo   SystemInfo `json:"sysinfo"`
}

// HeartbeatResponse is returned by hub
type HeartbeatResponse struct {
	OK              bool `json:"ok"`
	NextHeartbeatSec int  `json:"next_heartbeat_sec"`
}

// SystemInfo contains system metrics and information
type SystemInfo struct {
	Hostname    string   `json:"hostname"`
	OS          string   `json:"os"`
	OSVersion   string   `json:"os_version"`
	Arch        string   `json:"arch"`
	Uptime      int64    `json:"uptime"`
	CPUCount    int      `json:"cpu_count"`
	CPUUsage    float64  `json:"cpu_usage"`
	MemTotal    uint64   `json:"mem_total"`
	MemUsed     uint64   `json:"mem_used"`
	DiskTotal   uint64   `json:"disk_total"`
	DiskUsed    uint64   `json:"disk_used"`
	IPAddresses []string `json:"ip_addresses"`
	Timestamp   time.Time `json:"timestamp"`
}

// ErrorResponse represents API error
type ErrorResponse struct {
	Error string `json:"error"`
}
