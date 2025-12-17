package retry

import (
	"context"
	"sync"
	"time"
)

// State represents circuit breaker states
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastFailureTime  time.Time
	lastStateChange  time.Time
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures to open circuit
	SuccessThreshold int           // Number of successes to close from half-open
	Timeout          time.Duration // Time to wait before half-open
}

// DefaultCircuitBreakerConfig returns default circuit breaker config
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          1 * time.Minute,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: config.FailureThreshold,
		successThreshold: config.SuccessThreshold,
		timeout:          config.Timeout,
		lastStateChange:  time.Now(),
	}
}

// Call executes a function through the circuit breaker
func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	// Check if circuit should transition from open to half-open
	cb.checkStateTransition()

	// Check current state
	state := cb.getState()

	if state == StateOpen {
		return ErrCircuitOpen
	}

	// Execute function
	err := fn()

	// Record result
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// getState returns the current state
func (cb *CircuitBreaker) getState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetState returns the current state (public method)
func (cb *CircuitBreaker) GetState() State {
	return cb.getState()
}

// recordFailure records a failed call
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()
	cb.failureCount++

	if cb.state == StateHalfOpen {
		// Any failure in half-open state opens the circuit
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		cb.successCount = 0
	} else if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		// Too many failures, open the circuit
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}
}

// recordSuccess records a successful call
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			// Enough successes in half-open, close the circuit
			cb.state = StateClosed
			cb.lastStateChange = time.Now()
			cb.successCount = 0
		}
	}
}

// checkStateTransition checks if circuit should transition from open to half-open
func (cb *CircuitBreaker) checkStateTransition() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen && time.Since(cb.lastStateChange) >= cb.timeout {
		cb.state = StateHalfOpen
		cb.lastStateChange = time.Now()
		cb.successCount = 0
		cb.failureCount = 0
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateChange = time.Now()
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:           cb.state,
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		LastFailureTime: cb.lastFailureTime,
		LastStateChange: cb.lastStateChange,
	}
}

// CircuitBreakerStats holds circuit breaker statistics
type CircuitBreakerStats struct {
	State           State
	FailureCount    int
	SuccessCount    int
	LastFailureTime time.Time
	LastStateChange time.Time
}

// ErrCircuitOpen is returned when the circuit breaker is open
var ErrCircuitOpen = &CircuitBreakerError{Message: "circuit breaker is open"}

// CircuitBreakerError represents a circuit breaker error
type CircuitBreakerError struct {
	Message string
}

func (e *CircuitBreakerError) Error() string {
	return e.Message
}
