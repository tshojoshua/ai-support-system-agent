package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoff_Next(t *testing.T) {
	config := &Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.0, // No jitter for predictable testing
		MaxAttempts:  0,
	}

	backoff := NewBackoff(config)

	// First delay should be initial
	delay1 := backoff.Next()
	if delay1 != 100*time.Millisecond {
		t.Errorf("Expected first delay to be 100ms, got %v", delay1)
	}

	// Second delay should be doubled
	delay2 := backoff.Next()
	if delay2 != 200*time.Millisecond {
		t.Errorf("Expected second delay to be 200ms, got %v", delay2)
	}

	// Keep doubling until max
	backoff.Next() // 400ms
	backoff.Next() // 800ms
	delay5 := backoff.Next() // Should cap at 1s

	if delay5 != 1*time.Second {
		t.Errorf("Expected delay to cap at 1s, got %v", delay5)
	}
}

func TestBackoff_Reset(t *testing.T) {
	config := DefaultConfig()
	backoff := NewBackoff(config)

	backoff.Next()
	backoff.Next()
	
	if backoff.Attempts() != 2 {
		t.Errorf("Expected 2 attempts, got %d", backoff.Attempts())
	}

	backoff.Reset()

	if backoff.Attempts() != 0 {
		t.Errorf("Expected 0 attempts after reset, got %d", backoff.Attempts())
	}
}

func TestBackoff_ShouldContinue(t *testing.T) {
	config := &Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.0,
		MaxAttempts:  3,
	}

	backoff := NewBackoff(config)

	if !backoff.ShouldContinue() {
		t.Error("Should continue before any attempts")
	}

	backoff.Next() // Attempt 1
	backoff.Next() // Attempt 2
	backoff.Next() // Attempt 3

	if backoff.ShouldContinue() {
		t.Error("Should not continue after max attempts")
	}
}

func TestDo_Success(t *testing.T) {
	config := &Config{
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.0,
		MaxAttempts:  5,
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	ctx := context.Background()
	err := Do(ctx, config, fn)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestDo_ContextCanceled(t *testing.T) {
	config := &Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.0,
		MaxAttempts:  0, // Infinite
	}

	fn := func() error {
		return errors.New("always fails")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := Do(ctx, config, fn)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestExponential(t *testing.T) {
	tests := []struct {
		attempt      int
		initialDelay time.Duration
		maxDelay     time.Duration
		multiplier   float64
		expected     time.Duration
	}{
		{0, 100 * time.Millisecond, 1 * time.Second, 2.0, 100 * time.Millisecond},
		{1, 100 * time.Millisecond, 1 * time.Second, 2.0, 200 * time.Millisecond},
		{2, 100 * time.Millisecond, 1 * time.Second, 2.0, 400 * time.Millisecond},
		{10, 100 * time.Millisecond, 1 * time.Second, 2.0, 1 * time.Second}, // Capped at max
	}

	for _, tt := range tests {
		result := Exponential(tt.attempt, tt.initialDelay, tt.maxDelay, tt.multiplier)
		if result != tt.expected {
			t.Errorf("Exponential(%d, %v, %v, %f) = %v, want %v",
				tt.attempt, tt.initialDelay, tt.maxDelay, tt.multiplier, result, tt.expected)
		}
	}
}
