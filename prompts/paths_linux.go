//go:build linux

package config

func getDefaultStateDir() string {
	return "/var/lib/jtnt-agent"
}

func GetBinaryPath() string {
	return "/usr/local/bin/jtnt-agentd"
}
