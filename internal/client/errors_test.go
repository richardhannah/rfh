package client

import (
	"errors"
	"testing"
)

func TestRegistryError(t *testing.T) {
	t.Run("error with message", func(t *testing.T) {
		err := NewRegistryError(ErrPackageNotFound, "test package missing")
		expected := "package not found: test package missing"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error without message", func(t *testing.T) {
		err := NewRegistryError(ErrUnauthorized, "")
		expected := "unauthorized"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error unwrapping", func(t *testing.T) {
		err := NewRegistryError(ErrNetworkError, "connection timeout")
		if !errors.Is(err, ErrNetworkError) {
			t.Error("error should unwrap to ErrNetworkError")
		}
	})

	t.Run("error details initialization", func(t *testing.T) {
		err := NewRegistryError(ErrRateLimited, "too many requests")
		if err.Details == nil {
			t.Error("Details map should be initialized")
		}
		
		err.Details["retry_after"] = 60
		if err.Details["retry_after"] != 60 {
			t.Error("Details map should be writable")
		}
	})
}