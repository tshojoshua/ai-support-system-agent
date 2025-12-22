package api

import "time"

// EnrollRequest is sent to hub during enrollment
type EnrollRequest struct {
	Token         string                 `json:"token"`
	Hostname      string                 `json:"hostname"`
	OS            string                 `json:"os"`
	OSVersion     string                 `json:"os_version,omitempty"`
	Arch          string                 `json:"arch"`
	AgentVersion  string                 `json:"agent_version"`
	Capabilities  []string               `json:"capabilities"`
	SystemInfo    map[string]interface{} `json:"system_info,omitempty"`
}

// EnrollResponse is received from hub after successful enrollment
type EnrollResponse struct {
	AgentID              string `json:"agent_id"`
	AgentToken           string `json:"agent_token"`
	HubBaseURL           string `json:"hub_base_url"`
	PollIntervalSec      int    `json:"poll_interval_sec"`
	HeartbeatIntervalSec int    `json:"heartbeat_interval_sec"`
	TenantID             string `json:"tenant_id,omitempty"`
	SiteID               string `json:"site_id,omitempty"`
}

// HeartbeatRequest is sent periodically to hub
type HeartbeatRequest struct {
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time              `json:"timestamp"`
	Sysinfo   map[string]interface{} `json:"sysinfo"`
}

// HeartbeatResponse from hub
type HeartbeatResponse struct {
	OK               bool `json:"ok"`
	NextHeartbeatSec int  `json:"next_heartbeat_sec,omitempty"`
}
