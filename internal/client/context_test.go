package client

import (
	"context"
	"testing"
	"time"
)

func TestWithTimeout(t *testing.T) {
	t.Run("creates context with default timeout", func(t *testing.T) {
		ctx, cancel := WithTimeout(nil)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("context should have a deadline")
		}

		// Check that deadline is approximately DefaultTimeout from now
		expectedDeadline := time.Now().Add(DefaultTimeout)
		if deadline.Before(expectedDeadline.Add(-time.Second)) || deadline.After(expectedDeadline.Add(time.Second)) {
			t.Errorf("deadline should be approximately %v from now, got %v", DefaultTimeout, deadline.Sub(time.Now()))
		}
	})

	t.Run("uses existing context as parent", func(t *testing.T) {
		parentCtx := context.WithValue(context.Background(), "key", "value")
		ctx, cancel := WithTimeout(parentCtx)
		defer cancel()

		if ctx.Value("key") != "value" {
			t.Error("child context should inherit parent context values")
		}
	})
}

func TestWithCustomTimeout(t *testing.T) {
	t.Run("creates context with custom timeout", func(t *testing.T) {
		customTimeout := 5 * time.Second
		ctx, cancel := WithCustomTimeout(nil, customTimeout)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("context should have a deadline")
		}

		// Check that deadline is approximately customTimeout from now
		expectedDeadline := time.Now().Add(customTimeout)
		if deadline.Before(expectedDeadline.Add(-time.Second)) || deadline.After(expectedDeadline.Add(time.Second)) {
			t.Errorf("deadline should be approximately %v from now, got %v", customTimeout, deadline.Sub(time.Now()))
		}
	})

	t.Run("uses existing context as parent with custom timeout", func(t *testing.T) {
		parentCtx := context.WithValue(context.Background(), "test", "value")
		customTimeout := 10 * time.Second
		ctx, cancel := WithCustomTimeout(parentCtx, customTimeout)
		defer cancel()

		if ctx.Value("test") != "value" {
			t.Error("child context should inherit parent context values")
		}

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("context should have a deadline")
		}

		// Verify the custom timeout was applied
		expectedDeadline := time.Now().Add(customTimeout)
		if deadline.Before(expectedDeadline.Add(-time.Second)) || deadline.After(expectedDeadline.Add(time.Second)) {
			t.Errorf("deadline should be approximately %v from now, got %v", customTimeout, deadline.Sub(time.Now()))
		}
	})
}