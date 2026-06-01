package ai

import (
	"context"
	"time"
)

// withRetry executes fn with exponential backoff.
func withRetry(ctx context.Context, maxRetries int, initial time.Duration, fn func(context.Context) error) error {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if initial <= 0 {
		initial = 2 * time.Second
	}

	var err error
	backoff := initial
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err = fn(ctx); err == nil {
			return nil
		}
		if attempt == maxRetries {
			break
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		backoff *= 2
	}
	return err
}

func withTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, timeout)
}
