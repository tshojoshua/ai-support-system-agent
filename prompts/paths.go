package config

// Default paths - overridden by platform-specific files

var (
	stateDir  string
	configDir string
)

func GetStateDir() string {
	if stateDir == "" {
		stateDir = getDefaultStateDir()
	}
	return stateDir
}

func GetConfigPath() string {
	return GetStateDir() + "/config.json"
}

func GetLogsDir() string {
	return GetStateDir() + "/logs"
}
