package transport

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	initialBackoff = 30 * time.Second
	maxBackoff     = 15 * time.Minute
	jitterPercent  = 20 // ±20% jitter
)

// Backoff implements exponential backoff with jitter
type Backoff struct {
	attempt int
}

// NewBackoff creates a new backoff calculator
func NewBackoff() *Backoff {
	return &Backoff{attempt: 0}
}

// Next calculates the next backoff duration
func (b *Backoff) Next() time.Duration {
	b.attempt++
	return b.Calculate()
}

// Calculate returns the current backoff duration
func (b *Backoff) Calculate() time.Duration {
	// Exponential: initialBackoff * 2^attempt
	backoff := float64(initialBackoff) * math.Pow(2, float64(b.attempt-1))
	
	// Cap at maxBackoff
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}
	
	// Add jitter: ±20%
	jitter := backoff * (float64(jitterPercent) / 100.0)
	jitterAmount := (rand.Float64() * 2 * jitter) - jitter
	
	duration := time.Duration(backoff + jitterAmount)
	
	// Ensure minimum of initialBackoff
	if duration < initialBackoff {
		duration = initialBackoff
	}
	
	return duration
}

// Reset resets the backoff attempt counter
func (b *Backoff) Reset() {
	b.attempt = 0
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts int
	Backoff     *Backoff
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 0, // unlimited
		Backoff:     NewBackoff(),
	}
}

// ShouldRetry determines if an error is retryable
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// Retryable errors
	retryable := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"dial tcp",
		"dns",
		"temporary failure",
		"no such host",
		"network is unreachable",
		"i/o timeout",
	}
	
	for _, pattern := range retryable {
		if contains(errStr, pattern) {
			return true
		}
	}
	
	// Non-retryable errors
	nonRetryable := []string{
		"certificate",
		"tls",
		"x509",
		"bad request",
		"unauthorized",
		"forbidden",
	}
	
	for _, pattern := range nonRetryable {
		if contains(errStr, pattern) {
			return false
		}
	}
	
	// Default: retry unknown errors
	return true
}

// RetryWithBackoff executes fn with exponential backoff
func RetryWithBackoff(ctx context.Context, cfg *RetryConfig, fn func() error) error {
	var lastErr error
	
	for attempt := 1; cfg.MaxAttempts == 0 || attempt <= cfg.MaxAttempts; attempt++ {
		// Try operation
		if err := fn(); err == nil {
			cfg.Backoff.Reset()
			return nil
		} else {
			lastErr = err
			
			// Check if error is retryable
			if !ShouldRetry(err) {
				return fmt.Errorf("non-retryable error: %w", err)
			}
		}
		
		// Calculate backoff
		backoff := cfg.Backoff.Next()
		
		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}
	
	return fmt.Errorf("max attempts reached: %w", lastErr)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
