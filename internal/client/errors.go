package client

import "fmt"

// Common registry error types
var (
	ErrPackageNotFound   = fmt.Errorf("package not found")
	ErrVersionNotFound   = fmt.Errorf("version not found")
	ErrUnauthorized      = fmt.Errorf("unauthorized")
	ErrRateLimited       = fmt.Errorf("rate limited")
	ErrNetworkError      = fmt.Errorf("network error")
	ErrInvalidManifest   = fmt.Errorf("invalid manifest")
	ErrPublishFailed     = fmt.Errorf("publish failed")
	ErrConnectionFailed  = fmt.Errorf("connection failed")
	ErrInvalidRegistry   = fmt.Errorf("invalid registry")
	ErrNotImplemented    = fmt.Errorf("not implemented")
	ErrNotFound          = fmt.Errorf("not found")
	ErrInvalidOperation  = fmt.Errorf("invalid operation")
)

// RegistryError provides detailed error information
type RegistryError struct {
	Type    error
	Message string
	Details map[string]interface{}
}

func (e *RegistryError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%v: %s", e.Type, e.Message)
	}
	return e.Type.Error()
}

func (e *RegistryError) Unwrap() error {
	return e.Type
}

// NewRegistryError creates a new registry error
func NewRegistryError(errType error, message string) *RegistryError {
	return &RegistryError{
		Type:    errType,
		Message: message,
		Details: make(map[string]interface{}),
	}
}
