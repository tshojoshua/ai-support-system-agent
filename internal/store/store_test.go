package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_SaveLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "jtnt-store-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Test data
	key := "test-key"
	data := []byte("test data content")

	// Save
	if err := s.Save(key, data); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := s.Load(key)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Compare
	if string(loaded) != string(data) {
		t.Errorf("Data mismatch: got %s, want %s", loaded, data)
	}
}

func TestStore_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jtnt-store-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	key := "test-key"

	// Should not exist initially
	if s.Exists(key) {
		t.Error("Key should not exist before saving")
	}

	// Save
	if err := s.Save(key, []byte("data")); err != nil {
		t.Fatal(err)
	}

	// Should exist now
	if !s.Exists(key) {
		t.Error("Key should exist after saving")
	}
}

func TestStore_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jtnt-store-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	key := "test-key"

	// Save
	if err := s.Save(key, []byte("data")); err != nil {
		t.Fatal(err)
	}

	// Delete
	if err := s.Delete(key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should not exist
	if s.Exists(key) {
		t.Error("Key should not exist after deletion")
	}

	// Delete non-existent should not error
	if err := s.Delete(key); err != nil {
		t.Errorf("Delete() of non-existent key should not error: %v", err)
	}
}

func TestStore_Permissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jtnt-store-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	key := "test-key"
	data := []byte("sensitive data")

	if err := s.Save(key, data); err != nil {
		t.Fatal(err)
	}

	// Check file permissions (Unix only)
	if os.Getenv("GOOS") != "windows" {
		path := filepath.Join(tmpDir, key)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}

		// Should be 0600
		mode := info.Mode().Perm()
		expected := os.FileMode(0600)
		if mode != expected {
			t.Errorf("File permissions: got %o, want %o", mode, expected)
		}
	}
}
