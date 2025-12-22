//go:build windows

package store

func setSecurePermissions(path string) error {
	// On Windows, rely on default NTFS permissions
	// Files in ProgramData are already protected
	return nil
}
