// +build windows

package store

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SetPermissions sets secure permissions on Windows using ACLs
func (s *fileStore) SetPermissions(path string) error {
	// Set file permissions to owner only
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}
	
	// Get current user SID
	token, err := syscall.OpenCurrentProcessToken()
	if err != nil {
		return fmt.Errorf("failed to open process token: %w", err)
	}
	defer token.Close()
	
	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return fmt.Errorf("failed to get token user: %w", err)
	}
	
	// Convert path to UTF16
	pathUTF16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("failed to convert path: %w", err)
	}
	
	// Create DACL with only owner access
	// This is a simplified approach - production code should use proper ACL manipulation
	err = windows.SetNamedSecurityInfo(
		pathUTF16,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		tokenUser.User.Sid,
		nil,
		nil,
		nil,
	)
	
	if err != nil {
		return fmt.Errorf("failed to set security info: %w", err)
	}
	
	return nil
}

// secureDelete overwrites file before deletion (Windows-specific)
func secureDelete(path string) error {
	// Open file
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	
	// Get file size
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	
	// Overwrite with zeros
	zeros := make([]byte, stat.Size())
	if _, err := f.WriteAt(zeros, 0); err != nil {
		return err
	}
	
	f.Close()
	
	// Delete file
	return os.Remove(path)
}
