package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tshojoshua/jtnt-agent/internal/config"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

const (
	resultCacheDir = "job_results_pending"
	maxCacheAgeDays = 7
)

// ResultCache manages failed-to-upload job results
type ResultCache struct {
	cacheDir string
}

// NewResultCache creates a new result cache
func NewResultCache() (*ResultCache, error) {
	cacheDir := filepath.Join(config.GetStateDir(), resultCacheDir)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ResultCache{cacheDir: cacheDir}, nil
}

// Store saves a result to cache
func (rc *ResultCache) Store(jobID string, result *api.JobResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	filename := fmt.Sprintf("%s_%d.json", jobID, time.Now().Unix())
	path := filepath.Join(rc.cacheDir, filename)

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// List returns all cached results
func (rc *ResultCache) List() (map[string]*api.JobResult, error) {
	results := make(map[string]*api.JobResult)

	entries, err := os.ReadDir(rc.cacheDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(rc.cacheDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var result api.JobResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		results[path] = &result
	}

	return results, nil
}

// Delete removes a cached result
func (rc *ResultCache) Delete(path string) error {
	return os.Remove(path)
}

// Purge removes cached results older than maxAge
func (rc *ResultCache) Purge(maxAge time.Duration) error {
	entries, err := os.ReadDir(rc.cacheDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(rc.cacheDir, entry.Name())
			os.Remove(path)
		}
	}

	return nil
}

// ExtractJobID extracts job ID from cache filename
func ExtractJobID(filename string) string {
	base := filepath.Base(filename)
	// Format: jobID_timestamp.json
	if len(base) > 0 {
		// Find last underscore
		for i := len(base) - 1; i >= 0; i-- {
			if base[i] == '_' {
				return base[:i]
			}
		}
	}
	return base
}

// uploadCachedResults attempts to upload all cached results
func (a *Agent) uploadCachedResults(ctx context.Context) {
	if a.resultCache == nil {
		return
	}

	cached, err := a.resultCache.List()
	if err != nil {
		a.logger.Error("result-cache", map[string]interface{}{
			"message": "failed to list cached results",
			"error":   err.Error(),
		})
		return
	}

	if len(cached) == 0 {
		return
	}

	a.logger.Info("result-cache", map[string]interface{}{
		"message": "uploading cached results",
		"count":   len(cached),
	})

	for path, result := range cached {
		jobID := ExtractJobID(path)
		
		if err := a.jobExecutor.ReportResult(ctx, jobID, result); err != nil {
			a.logger.Error("result-cache", map[string]interface{}{
				"message": "failed to upload cached result",
				"job_id":  jobID,
				"error":   err.Error(),
			})
			continue
		}

		// Successfully uploaded, remove from cache
		if err := a.resultCache.Delete(path); err != nil {
			a.logger.Error("result-cache", map[string]interface{}{
				"message": "failed to delete cached result",
				"path":    path,
				"error":   err.Error(),
			})
		} else {
			a.logger.Info("result-cache", map[string]interface{}{
				"message": "uploaded and removed cached result",
				"job_id":  jobID,
			})
		}
	}

	// Purge old cached results
	if err := a.resultCache.Purge(maxCacheAgeDays * 24 * time.Hour); err != nil {
		a.logger.Error("result-cache", map[string]interface{}{
			"message": "failed to purge old cached results",
			"error":   err.Error(),
		})
	}
}
