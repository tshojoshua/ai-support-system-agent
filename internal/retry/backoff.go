package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// Config holds retry configuration
type Config struct {
	InitialDelay time.Duration // Initial delay between retries
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Backoff multiplier
	Jitter       float64       // Jitter factor (0-1)
	MaxAttempts  int           // Maximum attempts (0 = infinite)
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() *Config {
	return &Config{
		InitialDelay: 30 * time.Second,
		MaxDelay:     15 * time.Minute,
		Multiplier:   2.0,
		Jitter:       0.2,
		MaxAttempts:  0, // Infinite for heartbeat
	}
}

// NetworkOutageConfig returns config optimized for network outages
func NetworkOutageConfig() *Config {
	return &Config{
		InitialDelay: 1 * time.Minute,
		MaxDelay:     30 * time.Minute,
		Multiplier:   2.0,
		Jitter:       0.3,
		MaxAttempts:  0, // Infinite - survive 72-hour outages
	}
}

// Backoff implements exponential backoff with jitter
type Backoff struct {
	config       *Config
	attempt      int
	currentDelay time.Duration
	rng          *rand.Rand
}

// NewBackoff creates a new backoff instance
func NewBackoff(config *Config) *Backoff {
	if config == nil {
		config = DefaultConfig()
	}

	return &Backoff{
		config:       config,
		attempt:      0,
		currentDelay: config.InitialDelay,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Next returns the next backoff delay and increments the attempt counter
func (b *Backoff) Next() time.Duration {
	// Calculate base delay with exponential backoff
	delay := b.currentDelay

	// Add jitter
	if b.config.Jitter > 0 {
		jitterRange := float64(delay) * b.config.Jitter
		jitter := (b.rng.Float64() * 2 * jitterRange) - jitterRange
		delay = time.Duration(float64(delay) + jitter)
	}

	// Ensure delay is not negative
	if delay < 0 {
		delay = b.config.InitialDelay
	}

	// Cap at max delay
	if delay > b.config.MaxDelay {
		delay = b.config.MaxDelay
	}

	// Increment attempt
	b.attempt++

	// Calculate next delay for future use
	nextDelay := time.Duration(float64(b.currentDelay) * b.config.Multiplier)
	if nextDelay > b.config.MaxDelay {
		nextDelay = b.config.MaxDelay
	}
	b.currentDelay = nextDelay

	return delay
}

// Reset resets the backoff state
func (b *Backoff) Reset() {
	b.attempt = 0
	b.currentDelay = b.config.InitialDelay
}

// Attempts returns the number of attempts made
func (b *Backoff) Attempts() int {
	return b.attempt
}

// ShouldContinue returns true if more attempts should be made
func (b *Backoff) ShouldContinue() bool {
	if b.config.MaxAttempts == 0 {
		return true // Infinite retries
	}
	return b.attempt < b.config.MaxAttempts
}

// Do executes a function with retry logic
func Do(ctx context.Context, config *Config, fn func() error) error {
	backoff := NewBackoff(config)

	for {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if we should continue
		if !backoff.ShouldContinue() {
			return err
		}

		// Calculate next delay
		delay := backoff.Next()

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
}

// DoWithResult executes a function with retry logic and returns a result
func DoWithResult[T any](ctx context.Context, config *Config, fn func() (T, error)) (T, error) {
	backoff := NewBackoff(config)
	var zero T

	for {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		// Check if we should continue
		if !backoff.ShouldContinue() {
			return zero, err
		}

		// Calculate next delay
		delay := backoff.Next()

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
}

// Exponential calculates exponential backoff delay for attempt n
func Exponential(attempt int, initialDelay, maxDelay time.Duration, multiplier float64) time.Duration {
	delay := float64(initialDelay) * math.Pow(multiplier, float64(attempt))
	if delay > float64(maxDelay) {
		return maxDelay
	}
	return time.Duration(delay)
}

// WithJitter adds random jitter to a delay
func WithJitter(delay time.Duration, jitterFactor float64) time.Duration {
	if jitterFactor <= 0 || jitterFactor >= 1 {
		return delay
	}

	jitterRange := float64(delay) * jitterFactor
	jitter := (rand.Float64() * 2 * jitterRange) - jitterRange
	result := time.Duration(float64(delay) + jitter)

	if result < 0 {
		return delay
	}
	return result
}
