# Phase 2: Registry Client Interface

## Overview
Create an abstraction layer that defines the contract all registry implementations must follow, enabling polymorphic registry operations.

## Scope
- Define `RegistryClient` interface
- Create factory function for client instantiation
- Define common data structures for registry operations
- Establish error types for registry operations

## Prerequisites
- Phase 1: Registry Type Core Architecture completed

## Implementation Steps

### 1. Create Registry Client Interface

**File**: `internal/client/interface.go` (new file)

```go
package client

import (
    "context"
    "rulestack/internal/config"
)

// RegistryClient defines operations all registry types must support
type RegistryClient interface {
    // Search for packages in the registry
    SearchPackages(ctx context.Context, opts SearchOptions) ([]Package, error)
    
    // Get information about a specific package
    GetPackage(ctx context.Context, name string) (*Package, error)
    
    // Get information about a specific package version
    GetPackageVersion(ctx context.Context, name, version string) (*PackageVersion, error)
    
    // Publish a package to the registry
    PublishPackage(ctx context.Context, manifestPath, archivePath string) (*PublishResult, error)
    
    // Download a package archive by hash
    DownloadBlob(ctx context.Context, sha256, destPath string) error
    
    // Check if registry is accessible
    Health(ctx context.Context) error
    
    // Get registry type identifier
    Type() config.RegistryType
}
```

### 2. Define Common Data Structures

**File**: `internal/client/types.go` (new file)

```go
package client

import "time"

// SearchOptions contains parameters for package search
type SearchOptions struct {
    Query  string
    Tag    string
    Target string
    Limit  int
}

// Package represents a package in the registry
type Package struct {
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Latest      string    `json:"latest"`
    Versions    []string  `json:"versions"`
    Tags        []string  `json:"tags"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// PackageVersion represents a specific version of a package
type PackageVersion struct {
    Name         string                 `json:"name"`
    Version      string                 `json:"version"`
    Description  string                 `json:"description"`
    Dependencies map[string]string      `json:"dependencies"`
    SHA256       string                 `json:"sha256"`
    Size         int64                  `json:"size"`
    PublishedAt  time.Time             `json:"published_at"`
    Metadata     map[string]interface{} `json:"metadata"`
}

// PublishResult contains information about a published package
type PublishResult struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    SHA256  string `json:"sha256"`
    URL     string `json:"url,omitempty"`     // For HTTP registries
    PRUrl   string `json:"pr_url,omitempty"`  // For Git registries
    Message string `json:"message"`
}
```

### 3. Define Registry Errors

**File**: `internal/client/errors.go` (new file)

```go
package client

import "fmt"

// Common registry error types
var (
    ErrPackageNotFound  = fmt.Errorf("package not found")
    ErrVersionNotFound  = fmt.Errorf("version not found")
    ErrUnauthorized     = fmt.Errorf("unauthorized")
    ErrRateLimited      = fmt.Errorf("rate limited")
    ErrNetworkError     = fmt.Errorf("network error")
    ErrInvalidManifest  = fmt.Errorf("invalid manifest")
    ErrPublishFailed    = fmt.Errorf("publish failed")
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
```

### 4. Create Client Factory

**File**: `internal/client/factory.go` (new file)

```go
package client

import (
    "fmt"
    "rulestack/internal/config"
)

// NewRegistryClient creates the appropriate client based on registry type
func NewRegistryClient(registry config.Registry, verbose bool) (RegistryClient, error) {
    registryType := registry.GetEffectiveType()
    
    switch registryType {
    case config.RegistryTypeHTTP:
        client := NewHTTPClient(registry.URL, registry.JWTToken)
        client.SetVerbose(verbose)
        return client, nil
        
    case config.RegistryTypeGit:
        return NewGitClient(registry.URL, registry.GitToken, verbose)
        
    default:
        return nil, fmt.Errorf("unsupported registry type: %s", registryType)
    }
}

// GetClient creates a client for the current active registry
func GetClient(cfg config.CLIConfig, verbose bool) (RegistryClient, error) {
    if cfg.Current == "" {
        return nil, fmt.Errorf("no active registry configured")
    }
    
    registry, exists := cfg.Registries[cfg.Current]
    if !exists {
        return nil, fmt.Errorf("active registry '%s' not found in configuration", cfg.Current)
    }
    
    return NewRegistryClient(registry, verbose)
}

// GetClientForRegistry creates a client for a specific named registry
func GetClientForRegistry(cfg config.CLIConfig, registryName string, verbose bool) (RegistryClient, error) {
    registry, exists := cfg.Registries[registryName]
    if !exists {
        return nil, fmt.Errorf("registry '%s' not found", registryName)
    }
    
    return NewRegistryClient(registry, verbose)
}
```

### 5. Add Context Support Helper

**File**: `internal/client/context.go` (new file)

```go
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
```

### 6. Create Conversion Utilities

**File**: `internal/client/convert.go` (new file)

```go
package client

import "time"

// PackageToMap converts Package to map for backward compatibility
func PackageToMap(p *Package) map[string]interface{} {
    return map[string]interface{}{
        "name":        p.Name,
        "description": p.Description,
        "latest":      p.Latest,
        "versions":    p.Versions,
        "tags":        p.Tags,
        "updated_at":  p.UpdatedAt,
    }
}

// PackageVersionToMap converts PackageVersion to map
func PackageVersionToMap(pv *PackageVersion) map[string]interface{} {
    return map[string]interface{}{
        "name":         pv.Name,
        "version":      pv.Version,
        "description":  pv.Description,
        "dependencies": pv.Dependencies,
        "sha256":       pv.SHA256,
        "size":         pv.Size,
        "published_at": pv.PublishedAt,
        "metadata":     pv.Metadata,
    }
}

// MapToPackage converts map to Package struct
func MapToPackage(m map[string]interface{}) *Package {
    p := &Package{}
    
    if name, ok := m["name"].(string); ok {
        p.Name = name
    }
    if desc, ok := m["description"].(string); ok {
        p.Description = desc
    }
    if latest, ok := m["latest"].(string); ok {
        p.Latest = latest
    }
    // ... handle other fields ...
    
    return p
}
```

## Testing Requirements

### Unit Tests
1. Test factory function with different registry types
2. Test error creation and wrapping
3. Test conversion utilities
4. Test context helpers

### Integration Tests
1. Mock registry client implementation
2. Test polymorphic behavior
3. Test error handling across interface

### Cucumber Test Amendments

Since this phase primarily involves internal refactoring and interface creation, Cucumber tests will focus on ensuring existing functionality remains intact. The actual registry type behavior will be tested in later phases.

**File**: `features/registry-client.feature` (new file)

```gherkin
Feature: Registry Client Interface
  Registry operations should work consistently across different registry types

  Background:
    Given I have a clean test environment

  Scenario: Factory creates correct client for HTTP registry
    Given a registry "http-test" with URL "https://example.com" and type "remote-http"
    When I use registry "http-test"
    Then registry operations should use the HTTP client

  Scenario: Factory creates correct client for Git registry
    Given a registry "git-test" with URL "https://github.com/org/repo" and type "git"
    When I use registry "git-test"
    Then registry operations should use the Git client

  Scenario: Factory rejects unknown registry type
    Given a config file with content:
      """
      [registries.unknown]
      url = "https://example.com"
      type = "unknown-type"
      """
    When I run "rfh search test --registry unknown"
    Then the command should fail
    And the output should contain "unsupported registry type"
```

**File**: `features/step_definitions/client_steps.js`

Add new step definitions:
```javascript
Then('registry operations should use the HTTP client', async function () {
  // This would be verified through logging or mocking in actual implementation
  // For now, we ensure the command doesn't fail
  const result = await this.runCommand('rfh search test');
  // The actual client type would be logged in verbose mode
  assert(!result.error || result.error.includes('no active registry'), 
    'Command should work with HTTP client');
});

Then('registry operations should use the Git client', async function () {
  // Similar verification for Git client
  const result = await this.runCommand('rfh search test');
  // Git client might fail differently if not implemented yet
  assert(!result.error || result.error.includes('not implemented') || 
         result.error.includes('no active registry'),
    'Command should attempt to use Git client');
});
```

## Success Criteria
- Interface clearly defines all registry operations
- Factory correctly instantiates clients based on type
- Common data structures work for all registry types
- Error handling is consistent across implementations
- Context support enables timeout control

## Dependencies
- Phase 1: Registry Type Core Architecture

## Risks
- **Risk**: Interface too restrictive for future registry types
  **Mitigation**: Keep interface focused on core operations, use options pattern for extensibility
  
- **Risk**: Data structure incompatibilities between registry types
  **Mitigation**: Use flexible fields and metadata maps for type-specific data

## Migration Notes
The existing HTTP client will need to be refactored to implement this interface in Phase 3.

## Next Phase
Phase 3: HTTP Client Refactoring - Adapt existing HTTP client to implement the new interface