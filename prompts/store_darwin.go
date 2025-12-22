//go:build darwin

package store

import "os"

func setSecurePermissions(path string) error {
	// Set file to 0600 (owner read/write only)
	return os.Chmod(path, 0600)
}
