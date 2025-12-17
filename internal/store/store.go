package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// Store provides secure file storage interface
type Store interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
	Exists(key string) bool
	Delete(key string) error
	SetPermissions(path string) error
}

// NewStore creates a platform-specific store
func NewStore(baseDir string) (Store, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	
	return &fileStore{baseDir: baseDir}, nil
}

// fileStore implements Store using filesystem
type fileStore struct {
	baseDir string
}

func (s *fileStore) Save(key string, data []byte) error {
	path := s.getPath(key)
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	// Set platform-specific permissions
	if err := s.SetPermissions(path); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}
	
	return nil
}

func (s *fileStore) Load(key string) ([]byte, error) {
	path := s.getPath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

func (s *fileStore) Exists(key string) bool {
	path := s.getPath(key)
	_, err := os.Stat(path)
	return err == nil
}

func (s *fileStore) Delete(key string) error {
	path := s.getPath(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *fileStore) getPath(key string) string {
	return filepath.Join(s.baseDir, key)
}
