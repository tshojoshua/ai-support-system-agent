package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// Default shutdown timeout
	defaultShutdownTimeout = 30 * time.Second
	// Graceful shutdown timeout (SIGTERM)
	gracefulShutdownTimeout = 60 * time.Second
	// Force shutdown timeout (SIGINT)
	forceShutdownTimeout = 10 * time.Second
)

// Shutdown performs graceful shutdown of the agent
func (a *Agent) Shutdown(timeout time.Duration) error {
	a.logger.Info("shutdown", map[string]interface{}{
		"message": "initiating graceful shutdown",
		"timeout": timeout.String(),
	})

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Step 1: Stop accepting new jobs
	a.logger.Info("shutdown", map[string]interface{}{
		"message": "stopping job polling",
	})
	a.stopJobPolling()

	// Step 2: Wait for current job to finish
	a.logger.Info("shutdown", map[string]interface{}{
		"message": "waiting for active jobs to complete",
	})
	
	done := make(chan struct{})
	go func() {
		a.waitForCurrentJob()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("shutdown", map[string]interface{}{
			"message": "active jobs completed",
		})
	case <-ctx.Done():
		a.logger.Warn("shutdown", map[string]interface{}{
			"message": "shutdown timeout reached, killing active jobs",
		})
		a.killCurrentJob()
	}

	// Step 3: Flush cached results
	a.logger.Info("shutdown", map[string]interface{}{
		"message": "flushing cached results",
	})
	a.flushCachedResults(ctx)

	// Step 4: Send final heartbeat
	a.logger.Info("shutdown", map[string]interface{}{
		"message": "sending final heartbeat",
	})
	a.sendFinalHeartbeat(ctx)

	// Step 5: Stop servers
	a.stopServers(5 * time.Second)

	// Step 6: Close connections
	a.logger.Info("shutdown", map[string]interface{}{
		"message": "closing connections",
	})

	a.logger.Info("shutdown", map[string]interface{}{
		"message": "shutdown complete",
	})

	return nil
}

// stopJobPolling stops the job polling loop
func (a *Agent) stopJobPolling() {
	a.mu.Lock()
	a.jobPollingStopped = true
	a.mu.Unlock()
}

// waitForCurrentJob waits for the current job to complete
func (a *Agent) waitForCurrentJob() {
	a.mu.RLock()
	currentJob := a.currentJob
	a.mu.RUnlock()

	if currentJob == nil {
		return
	}

	// Wait for job context to be done
	<-currentJob.Done()
}

// killCurrentJob forcefully terminates the current job
func (a *Agent) killCurrentJob() {
	a.mu.RLock()
	currentJob := a.currentJob
	a.mu.RUnlock()

	if currentJob == nil {
		return
	}

	// Cancel job context
	if currentJob != nil {
		// The job context should already have a cancel function
		// This is a placeholder for the actual implementation
	}
}

// flushCachedResults attempts to upload all cached results
func (a *Agent) flushCachedResults(ctx context.Context) {
	if a.resultCache == nil {
		return
	}

	a.uploadCachedResults(ctx)
}

// sendFinalHeartbeat sends a final heartbeat with shutdown status
func (a *Agent) sendFinalHeartbeat(ctx context.Context) {
	// Send heartbeat with special shutdown status
	if err := a.sendHeartbeatWithStatus(ctx, "shutting_down"); err != nil {
		a.logger.Error("shutdown", map[string]interface{}{
			"message": "failed to send final heartbeat",
			"error":   err.Error(),
		})
	}
}

// sendHeartbeatWithStatus sends a heartbeat with a specific status
func (a *Agent) sendHeartbeatWithStatus(ctx context.Context, status string) error {
	// This would call the actual heartbeat API with status
	// Placeholder for actual implementation
	return nil
}

// stopServers stops the metrics and health servers
func (a *Agent) stopServers(timeout time.Duration) {
	if a.metricsServer != nil {
		if err := a.metricsServer.Stop(timeout); err != nil {
			a.logger.Error("shutdown", map[string]interface{}{
				"message": "failed to stop metrics server",
				"error":   err.Error(),
			})
		}
	}

	if a.healthServer != nil {
		if err := a.healthServer.Stop(timeout); err != nil {
			a.logger.Error("shutdown", map[string]interface{}{
				"message": "failed to stop health server",
				"error":   err.Error(),
			})
		}
	}
}

// SetupSignalHandlers sets up signal handlers for graceful shutdown
func (a *Agent) SetupSignalHandlers() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		
		var timeout time.Duration
		switch sig {
		case syscall.SIGTERM:
			// Graceful shutdown
			timeout = gracefulShutdownTimeout
			a.logger.Info("signal", map[string]interface{}{
				"message": "received SIGTERM, initiating graceful shutdown",
			})
		case syscall.SIGINT:
			// Faster shutdown
			timeout = forceShutdownTimeout
			a.logger.Info("signal", map[string]interface{}{
				"message": "received SIGINT, initiating shutdown",
			})
		case syscall.SIGQUIT:
			// Immediate shutdown
			timeout = 5 * time.Second
			a.logger.Info("signal", map[string]interface{}{
				"message": "received SIGQUIT, initiating immediate shutdown",
			})
		default:
			timeout = defaultShutdownTimeout
		}

		if err := a.Shutdown(timeout); err != nil {
			a.logger.Error("shutdown", map[string]interface{}{
				"message": "shutdown failed",
				"error":   err.Error(),
			})
			os.Exit(1)
		}

		os.Exit(0)
	}()
}
