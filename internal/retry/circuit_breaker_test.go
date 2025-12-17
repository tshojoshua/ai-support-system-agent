package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	}

	cb := NewCircuitBreaker(config)

	if cb.GetState() != StateClosed {
		t.Error("Circuit should start in closed state")
	}

	// Trigger failures
	for i := 0; i < 3; i++ {
		cb.Call(context.Background(), func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != StateOpen {
		t.Error("Circuit should be open after threshold failures")
	}
}

func TestCircuitBreaker_OpenToHalfOpen(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}

	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.Call(context.Background(), func() error { return errors.New("fail") })
	cb.Call(context.Background(), func() error { return errors.New("fail") })

	if cb.GetState() != StateOpen {
		t.Error("Circuit should be open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to half-open
	cb.Call(context.Background(), func() error { return nil })

	if cb.GetState() != StateHalfOpen {
		t.Errorf("Circuit should be half-open, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenToClosed(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}

	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.Call(context.Background(), func() error { return errors.New("fail") })
	cb.Call(context.Background(), func() error { return errors.New("fail") })

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Successful calls to close
	cb.Call(context.Background(), func() error { return nil })
	cb.Call(context.Background(), func() error { return nil })

	if cb.GetState() != StateClosed {
		t.Errorf("Circuit should be closed, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}

	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.Call(context.Background(), func() error { return errors.New("fail") })
	cb.Call(context.Background(), func() error { return errors.New("fail") })

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Failure in half-open should open again
	cb.Call(context.Background(), func() error { return errors.New("fail") })

	if cb.GetState() != StateOpen {
		t.Errorf("Circuit should be open, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_OpenRejectsCall(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
	}

	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.Call(context.Background(), func() error { return errors.New("fail") })

	// Call should be rejected
	err := cb.Call(context.Background(), func() error { return nil })

	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	// Open the circuit
	for i := 0; i < 5; i++ {
		cb.Call(context.Background(), func() error { return errors.New("fail") })
	}

	if cb.GetState() != StateOpen {
		t.Error("Circuit should be open")
	}

	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Error("Circuit should be closed after reset")
	}

	stats := cb.GetStats()
	if stats.FailureCount != 0 || stats.SuccessCount != 0 {
		t.Error("Stats should be reset")
	}
}
