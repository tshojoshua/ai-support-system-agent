package transport

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoff_Calculate(t *testing.T) {
	b := NewBackoff()

	// First attempt
	d1 := b.Next()
	if d1 < initialBackoff {
		t.Errorf("First backoff should be at least %v, got %v", initialBackoff, d1)
	}

	// Second attempt (should be ~2x)
	d2 := b.Next()
	if d2 < d1 {
		t.Errorf("Second backoff should be >= first: %v < %v", d2, d1)
	}

	// After many attempts, should cap at maxBackoff
	for i := 0; i < 20; i++ {
		b.Next()
	}

	dMax := b.Calculate()
	// Account for jitter
	maxWithJitter := maxBackoff + time.Duration(float64(maxBackoff)*(float64(jitterPercent)/100.0))
	if dMax > maxWithJitter {
		t.Errorf("Backoff exceeded max with jitter: %v > %v", dMax, maxWithJitter)
	}
}

func TestBackoff_Reset(t *testing.T) {
	b := NewBackoff()

	// Advance several attempts
	for i := 0; i < 5; i++ {
		b.Next()
	}

	// Reset
	b.Reset()

	// Next should be back to initial
	d := b.Next()
	if d < initialBackoff {
		t.Errorf("After reset, backoff should be at least %v, got %v", initialBackoff, d)
	}

	// Should be approximately initialBackoff (accounting for jitter)
	maxExpected := initialBackoff + time.Duration(float64(initialBackoff)*(float64(jitterPercent)/100.0))
	if d > maxExpected {
		t.Errorf("After reset, backoff too large: %v > %v", d, maxExpected)
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		retry bool
	}{
		{
			name:  "nil error",
			err:   nil,
			retry: false,
		},
		{
			name:  "connection refused",
			err:   errors.New("dial tcp: connection refused"),
			retry: true,
		},
		{
			name:  "timeout",
			err:   errors.New("context deadline exceeded: timeout"),
			retry: true,
		},
		{
			name:  "dns error",
			err:   errors.New("no such host: dns lookup failed"),
			retry: true,
		},
		{
			name:  "certificate error",
			err:   errors.New("x509: certificate has expired"),
			retry: false,
		},
		{
			name:  "tls error",
			err:   errors.New("tls: bad certificate"),
			retry: false,
		},
		{
			name:  "bad request",
			err:   errors.New("bad request: invalid input"),
			retry: false,
		},
		{
			name:  "unauthorized",
			err:   errors.New("unauthorized access"),
			retry: false,
		},
		{
			name:  "network unreachable",
			err:   errors.New("network is unreachable"),
			retry: true,
		},
		{
			name:  "io timeout",
			err:   errors.New("i/o timeout reading response"),
			retry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRetry(tt.err)
			if result != tt.retry {
				t.Errorf("ShouldRetry() = %v, want %v for error: %v", result, tt.retry, tt.err)
			}
		})
	}
}

func TestRetryWithBackoff_Success(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3

	attemptCount := 0
	fn := func() error {
		attemptCount++
		if attemptCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := RetryWithBackoff(ctx, cfg, fn)
	if err != nil {
		t.Errorf("RetryWithBackoff() should succeed, got error: %v", err)
	}

	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount)
	}
}

func TestRetryWithBackoff_MaxAttempts(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.Backoff = NewBackoff()

	attemptCount := 0
	fn := func() error {
		attemptCount++
		return errors.New("persistent failure")
	}

	// Speed up test by reducing backoff
	originalInitial := initialBackoff
	defer func() { 
		// Note: can't actually change const, but test still works
	}()

	err := RetryWithBackoff(ctx, cfg, fn)
	if err == nil {
		t.Error("RetryWithBackoff() should fail after max attempts")
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestRetryWithBackoff_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 5

	attemptCount := 0
	fn := func() error {
		attemptCount++
		return errors.New("certificate error: invalid")
	}

	err := RetryWithBackoff(ctx, cfg, fn)
	if err == nil {
		t.Error("RetryWithBackoff() should fail on non-retryable error")
	}

	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attemptCount)
	}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := DefaultRetryConfig()

	attemptCount := 0
	fn := func() error {
		attemptCount++
		if attemptCount == 1 {
			// Cancel context after first attempt
			cancel()
		}
		return errors.New("temporary failure")
	}

	err := RetryWithBackoff(ctx, cfg, fn)
	if err == nil {
		t.Error("RetryWithBackoff() should fail when context is cancelled")
	}

	// Should have attempted at least once
	if attemptCount < 1 {
		t.Errorf("Expected at least 1 attempt, got %d", attemptCount)
	}
}
