// +build !windows,!darwin

package store

import (
	"fmt"
	"os"
	"syscall"
)

// SetPermissions sets secure permissions on Linux
func (s *fileStore) SetPermissions(path string) error {
	// Set file permissions to 0600 (owner read/write only)
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}
	
	// Ensure file is owned by root or current user
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	
	sysstat, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to get file system stats")
	}
	
	currentUID := os.Getuid()
	if sysstat.Uid != uint32(currentUID) && sysstat.Uid != 0 {
		return fmt.Errorf("file has unexpected owner")
	}
	
	return nil
}
