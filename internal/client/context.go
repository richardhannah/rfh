package client

import (
	"context"
	"time"
)

// DefaultTimeout is the default timeout for registry operations
const DefaultTimeout = 30 * time.Second

// WithTimeout creates a context with the default timeout
func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, DefaultTimeout)
}

// WithCustomTimeout creates a context with a custom timeout
func WithCustomTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, timeout)
}
