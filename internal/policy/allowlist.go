package policy

import (
	"path/filepath"
	"strings"
)

// Allowlist provides path and command matching
type Allowlist struct {
	patterns []string
}

// NewAllowlist creates a new allowlist from patterns
func NewAllowlist(patterns []string) *Allowlist {
	return &Allowlist{patterns: patterns}
}

// Allows checks if a path/command is allowed by any pattern
func (a *Allowlist) Allows(path string) bool {
	if len(a.patterns) == 0 {
		return false
	}

	// Normalize path separators
	path = filepath.Clean(path)

	for _, pattern := range a.patterns {
		if a.matchPattern(pattern, path) {
			return true
		}
	}

	return false
}

// matchPattern checks if path matches glob pattern
func (a *Allowlist) matchPattern(pattern, path string) bool {
	// Normalize pattern
	pattern = filepath.Clean(pattern)

	// Exact match
	if pattern == path {
		return true
	}

	// Glob match
	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}

	// Directory prefix match for patterns ending with /*
	if strings.HasSuffix(pattern, string(filepath.Separator)+"*") {
		dir := strings.TrimSuffix(pattern, string(filepath.Separator)+"*")
		if strings.HasPrefix(path, dir+string(filepath.Separator)) {
			return true
		}
		if path == dir {
			return true
		}
	}

	return false
}

// AllowsBinary checks if a binary name is allowed
func AllowsBinary(allowedBinaries []string, binary string) bool {
	// Extract binary name without path
	binaryName := filepath.Base(binary)

	// Remove extension on Windows
	if ext := filepath.Ext(binaryName); ext == ".exe" || ext == ".cmd" || ext == ".bat" {
		binaryName = strings.TrimSuffix(binaryName, ext)
	}

	for _, allowed := range allowedBinaries {
		allowedName := filepath.Base(allowed)
		if ext := filepath.Ext(allowedName); ext == ".exe" || ext == ".cmd" || ext == ".bat" {
			allowedName = strings.TrimSuffix(allowedName, ext)
		}

		if strings.EqualFold(binaryName, allowedName) {
			return true
		}
	}

	return false
}

// ValidatePath checks if path is safe (no traversal attacks)
func ValidatePath(path string) error {
	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for path traversal
	if strings.Contains(cleaned, "..") {
		return ErrPathTraversal
	}

	// Check for suspicious patterns
	suspicious := []string{
		"/../",
		"\\..\\",
		"/./",
		"\\.\\",
	}

	for _, pattern := range suspicious {
		if strings.Contains(path, pattern) {
			return ErrPathTraversal
		}
	}

	return nil
}
