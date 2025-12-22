//go:build windows

package config

import (
	"os"
	"path/filepath"
)

func getDefaultStateDir() string {
	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = "C:\\ProgramData"
	}
	return filepath.Join(programData, "JTNT", "Agent")
}

func GetBinaryPath() string {
	programFiles := os.Getenv("PROGRAMFILES")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	return filepath.Join(programFiles, "JTNT", "Agent", "jtnt-agentd.exe")
}
