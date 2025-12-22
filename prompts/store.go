package store

import (
	"fmt"
	"os"
)

type Store interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
	Exists(key string) bool
	Delete(key string) error
}

type fileStore struct {
	baseDir string
}

func New(baseDir string) Store {
	return &fileStore{baseDir: baseDir}
}

func (s *fileStore) Save(key string, data []byte) error {
	path := s.baseDir + "/" + key

	// Ensure directory exists
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file with secure permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Platform-specific permission setting
	if err := setSecurePermissions(path); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

func (s *fileStore) Load(key string) ([]byte, error) {
	path := s.baseDir + "/" + key
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

func (s *fileStore) Exists(key string) bool {
	path := s.baseDir + "/" + key
	_, err := os.Stat(path)
	return err == nil
}

func (s *fileStore) Delete(key string) error {
	path := s.baseDir + "/" + key
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
