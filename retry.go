package vision

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

const (
	defaultMaxRetries = 2
	retryDelay        = 1 * time.Second
)

// RetryResolver wraps a Resolver with retry logic on transient failures.
type RetryResolver struct {
	inner      Resolver
	maxRetries int
}

// NewRetryResolver creates a new RetryResolver wrapping the given resolver.
// If maxRetries <= 0, defaultMaxRetries is used.
func NewRetryResolver(inner Resolver, maxRetries int) *RetryResolver {
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}
	return &RetryResolver{
		inner:      inner,
		maxRetries: maxRetries,
	}
}

// Resolve delegates to the inner resolver with retry on ErrServiceUnavailable.
func (r *RetryResolver) Resolve(ctx context.Context, locate string, screenshot []byte) (Coordinates, error) {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return Coordinates{}, ctx.Err()
		}

		if attempt > 0 {
			slog.WarnContext(ctx, "retrying vision resolve",
				"locate", locate,
				"attempt", attempt+1,
			)
			time.Sleep(retryDelay)
		}

		coords, err := r.inner.Resolve(ctx, locate, screenshot)
		if err == nil {
			return coords, nil
		}

		lastErr = err

		if !errors.Is(err, ErrServiceUnavailable) {
			break
		}

		slog.WarnContext(ctx, "vision resolve attempt failed",
			"locate", locate,
			"attempt", attempt+1,
			"error", err,
		)
	}

	if errors.Is(lastErr, ErrServiceUnavailable) {
		return Coordinates{}, fmt.Errorf("vision service temporarily unavailable: %w", lastErr)
	}

	return Coordinates{}, fmt.Errorf("%w: cannot locate element '%s': %v", ErrElementNotFound, locate, lastErr)
}
