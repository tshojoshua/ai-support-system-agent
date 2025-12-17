package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

const (
	defaultJobPollInterval   = 30 * time.Second
	minJobPollInterval       = 10 * time.Second
	maxJobPollInterval       = 5 * time.Minute
	cacheUploadInterval      = 5 * time.Minute
	errorBackoffMultiplier   = 2.0
	maxErrorBackoff          = 5 * time.Minute
)

// jobPollLoop continuously polls for jobs from hub and executes them
func (a *Agent) jobPollLoop(ctx context.Context) {
	pollInterval := defaultJobPollInterval
	errorBackoff := minJobPollInterval
	lastCacheUpload := time.Now()

	jobTicker := time.NewTicker(pollInterval)
	defer jobTicker.Stop()

	a.logger.Info("job-poll", map[string]interface{}{
		"message":  "job polling loop started",
		"interval": pollInterval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("job-poll", map[string]interface{}{
				"message": "job polling loop stopped",
			})
			return

		case <-jobTicker.C:
			// Attempt to upload cached results periodically
			if time.Since(lastCacheUpload) >= cacheUploadInterval {
				a.uploadCachedResults(ctx)
				lastCacheUpload = time.Now()
			}

			// Fetch and execute next job
			if err := a.processNextJob(ctx); err != nil {
				a.logger.Error("job-poll", map[string]interface{}{
					"message": "job processing error",
					"error":   err.Error(),
				})

				// Apply exponential backoff on errors
				errorBackoff = time.Duration(float64(errorBackoff) * errorBackoffMultiplier)
				if errorBackoff > maxErrorBackoff {
					errorBackoff = maxErrorBackoff
				}

				jobTicker.Reset(errorBackoff)
				a.logger.Debug("job-poll", map[string]interface{}{
					"message":       "applying error backoff",
					"next_poll_in":  errorBackoff.String(),
				})
			} else {
				// Reset backoff on success
				errorBackoff = minJobPollInterval
				jobTicker.Reset(pollInterval)
			}
		}
	}
}

// processNextJob fetches and executes a single job
func (a *Agent) processNextJob(ctx context.Context) error {
	// Fetch next job from hub
	job, err := a.jobExecutor.FetchNextJob(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch job: %w", err)
	}

	// No job available
	if job == nil {
		a.logger.Debug("job-poll", map[string]interface{}{
			"message": "no jobs available",
		})
		return nil
	}

	a.logger.Info("job-execute", map[string]interface{}{
		"message": "job received",
		"job_id":  job.ID,
		"type":    job.Type,
	})

	// Execute job with timeout context
	execCtx := ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(job.Timeout)*time.Second)
		defer cancel()
	}

	result := a.jobExecutor.Execute(execCtx, job)

	a.logger.Info("job-execute", map[string]interface{}{
		"message":    "job completed",
		"job_id":     job.ID,
		"status":     result.Status,
		"exit_code":  result.ExitCode,
	})

	// Report result to hub
	if err := a.jobExecutor.ReportResult(ctx, job.ID, result); err != nil {
		a.logger.Error("job-execute", map[string]interface{}{
			"message": "failed to report job result",
			"job_id":  job.ID,
			"error":   err.Error(),
		})

		// Cache result for later upload
		if a.resultCache != nil {
			if cacheErr := a.resultCache.Store(job.ID, result); cacheErr != nil {
				a.logger.Error("job-execute", map[string]interface{}{
					"message": "failed to cache job result",
					"job_id":  job.ID,
					"error":   cacheErr.Error(),
				})
			} else {
				a.logger.Info("job-execute", map[string]interface{}{
					"message": "cached job result for later upload",
					"job_id":  job.ID,
				})
			}
		}

		return fmt.Errorf("failed to report result: %w", err)
	}

	a.logger.Info("job-execute", map[string]interface{}{
		"message": "job result reported successfully",
		"job_id":  job.ID,
	})

	return nil
}

// updatePollInterval updates the job polling interval from hub configuration
func (a *Agent) updatePollInterval(interval int) time.Duration {
	if interval <= 0 {
		return defaultJobPollInterval
	}

	newInterval := time.Duration(interval) * time.Second

	if newInterval < minJobPollInterval {
		return minJobPollInterval
	}

	if newInterval > maxJobPollInterval {
		return maxJobPollInterval
	}

	return newInterval
}
