package transport

import (
	"context"
	"math/rand"
	"time"
)

type RetryConfig struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
	MaxAttempts  int
}

var DefaultRetryConfig = RetryConfig{
	InitialDelay: 30 * time.Second,
	MaxDelay:     15 * time.Minute,
	Multiplier:   2.0,
	Jitter:       0.2,
	MaxAttempts:  0, // Infinite for heartbeat
}

func WithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	attempt := 0
	delay := cfg.InitialDelay

	for {
		err := fn()
		if err == nil {
			return nil
		}

		attempt++
		if cfg.MaxAttempts > 0 && attempt >= cfg.MaxAttempts {
			return err
		}

		// Calculate next delay with exponential backoff
		jitter := 1.0 + (rand.Float64()*2.0-1.0)*cfg.Jitter
		nextDelay := time.Duration(float64(delay) * cfg.Multiplier * jitter)
		if nextDelay > cfg.MaxDelay {
			nextDelay = cfg.MaxDelay
		}

		select {
		case <-time.After(delay):
			delay = nextDelay
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
