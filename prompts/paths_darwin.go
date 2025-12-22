//go:build darwin

package config

func getDefaultStateDir() string {
	return "/Library/Application Support/JTNT/Agent"
}

func GetBinaryPath() string {
	return "/usr/local/jtnt/agent/jtnt-agentd"
}
