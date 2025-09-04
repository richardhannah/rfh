# Phase 3: HTTP Client Refactoring

## Overview
Refactor the existing HTTP client to implement the new `RegistryClient` interface, maintaining backward compatibility while enabling polymorphic registry operations.

## Scope
- Rename existing client.go to http.go
- Implement RegistryClient interface
- Add context support to all operations
- Convert between old map-based and new struct-based data formats
- Update all CLI commands to use the new interface

## Prerequisites
- Phase 1: Registry Type Core Architecture completed
- Phase 2: Registry Client Interface completed

## Implementation Steps

### 1. Rename and Update HTTP Client

**File**: Rename `internal/client/client.go` to `internal/client/http.go`

```go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
    
    "rulestack/internal/config"
)

// HTTPClient represents an HTTP client for the RuleStack registry
type HTTPClient struct {
    baseURL    string
    token      string
    httpClient *http.Client
    verbose    bool
}

// Ensure HTTPClient implements RegistryClient
var _ RegistryClient = (*HTTPClient)(nil)

// NewHTTPClient creates a new HTTP registry client
func NewHTTPClient(baseURL, token string) *HTTPClient {
    baseURL = strings.TrimRight(baseURL, "/")
    
    return &HTTPClient{
        baseURL: baseURL,
        token:   token,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        verbose: false,
    }
}

// Type returns the registry type
func (c *HTTPClient) Type() config.RegistryType {
    return config.RegistryTypeHTTP
}
```

### 2. Update Search Method

**File**: `internal/client/http.go`

```go
// SearchPackages searches for packages in the registry
func (c *HTTPClient) SearchPackages(ctx context.Context, opts SearchOptions) ([]Package, error) {
    path := "/v1/packages"
    
    // Build query parameters
    params := url.Values{}
    if opts.Query != "" {
        params.Add("q", opts.Query)
    }
    if opts.Tag != "" {
        params.Add("tag", opts.Tag)
    }
    if opts.Target != "" {
        params.Add("target", opts.Target)
    }
    if opts.Limit > 0 {
        params.Add("limit", strconv.Itoa(opts.Limit))
    }
    
    if len(params) > 0 {
        path += "?" + params.Encode()
    }
    
    resp, err := c.makeRequestWithContext(ctx, "GET", path, nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, NewRegistryError(ErrNetworkError, 
            fmt.Sprintf("search failed (status %d): %s", resp.StatusCode, string(body)))
    }
    
    // Parse response as maps first (for backward compatibility)
    var results []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    // Convert to Package structs
    packages := make([]Package, len(results))
    for i, r := range results {
        packages[i] = *MapToPackage(r)
    }
    
    return packages, nil
}
```

### 3. Update Get Methods

**File**: `internal/client/http.go`

```go
// GetPackage gets information about a specific package
func (c *HTTPClient) GetPackage(ctx context.Context, name string) (*Package, error) {
    path := fmt.Sprintf("/v1/packages/%s", name)
    
    resp, err := c.makeRequestWithContext(ctx, "GET", path, nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusNotFound {
        return nil, NewRegistryError(ErrPackageNotFound, name)
    }
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, NewRegistryError(ErrNetworkError,
            fmt.Sprintf("request failed (status %d): %s", resp.StatusCode, string(body)))
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return MapToPackage(result), nil
}

// GetPackageVersion gets information about a specific package version
func (c *HTTPClient) GetPackageVersion(ctx context.Context, name, version string) (*PackageVersion, error) {
    path := fmt.Sprintf("/v1/packages/%s/versions/%s", name, version)
    
    resp, err := c.makeRequestWithContext(ctx, "GET", path, nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusNotFound {
        return nil, NewRegistryError(ErrVersionNotFound, 
            fmt.Sprintf("%s@%s", name, version))
    }
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, NewRegistryError(ErrNetworkError,
            fmt.Sprintf("request failed (status %d): %s", resp.StatusCode, string(body)))
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return MapToPackageVersion(result), nil
}
```

### 4. Update Publish Method

**File**: `internal/client/http.go`

```go
// PublishPackage publishes a package to the registry
func (c *HTTPClient) PublishPackage(ctx context.Context, manifestPath, archivePath string) (*PublishResult, error) {
    // Create multipart form
    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)
    
    // Add manifest file
    if err := c.addFileToForm(writer, "manifest", manifestPath); err != nil {
        return nil, fmt.Errorf("failed to add manifest: %w", err)
    }
    
    // Add archive file
    if err := c.addFileToForm(writer, "archive", archivePath); err != nil {
        return nil, fmt.Errorf("failed to add archive: %w", err)
    }
    
    writer.Close()
    
    // Make request
    resp, err := c.makeRequestWithContext(ctx, "POST", "/v1/packages", &buf, writer.FormDataContentType())
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    
    if resp.StatusCode == http.StatusUnauthorized {
        return nil, NewRegistryError(ErrUnauthorized, "authentication required")
    }
    
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return nil, NewRegistryError(ErrPublishFailed,
            fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &PublishResult{
        Name:    result["name"].(string),
        Version: result["version"].(string),
        SHA256:  result["sha256"].(string),
        URL:     c.baseURL + path,
        Message: "Package published successfully",
    }, nil
}
```

### 5. Add Context Support to Request Method

**File**: `internal/client/http.go`

```go
// makeRequestWithContext makes an HTTP request with authentication and context
func (c *HTTPClient) makeRequestWithContext(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
    url := c.baseURL + path
    
    if c.verbose {
        fmt.Printf("🌐 %s %s\n", method, url)
    }
    
    req, err := http.NewRequestWithContext(ctx, method, url, body)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add authentication header
    if c.token != "" {
        authHeader := "Bearer " + c.token
        req.Header.Set("Authorization", authHeader)
        if c.verbose {
            tokenPreview := c.token
            if len(tokenPreview) > 20 {
                tokenPreview = tokenPreview[:20] + "..."
            }
            fmt.Printf("🔍 Setting Authorization header: Bearer %s\n", tokenPreview)
        }
    }
    
    // Set content type if provided
    if contentType != "" {
        req.Header.Set("Content-Type", contentType)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        if ctx.Err() != nil {
            return nil, fmt.Errorf("request canceled: %w", ctx.Err())
        }
        return nil, fmt.Errorf("request failed: %w", err)
    }
    
    if c.verbose {
        fmt.Printf("🔍 HTTP Response: %d %s\n", resp.StatusCode, resp.Status)
    }
    
    return resp, nil
}
```

### 6. Update CLI Commands

**File**: `internal/cli/search.go` (example update)

```go
func runSearch(query, tag, target string, limit int) error {
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Use new factory function
    client, err := client.GetClient(cfg, verboseFlag)
    if err != nil {
        return err
    }
    
    ctx, cancel := client.WithTimeout(context.Background())
    defer cancel()
    
    // Use new search options
    opts := client.SearchOptions{
        Query:  query,
        Tag:    tag,
        Target: target,
        Limit:  limit,
    }
    
    packages, err := client.SearchPackages(ctx, opts)
    if err != nil {
        return fmt.Errorf("search failed: %w", err)
    }
    
    // Display results using new Package struct
    for _, pkg := range packages {
        fmt.Printf("📦 %s@%s\n", pkg.Name, pkg.Latest)
        if pkg.Description != "" {
            fmt.Printf("   %s\n", pkg.Description)
        }
        // ... rest of display logic
    }
    
    return nil
}
```

### 7. Create Backward Compatibility Layer

**File**: `internal/client/compat.go` (new file)

```go
package client

// LegacyClient wraps the new client interface for backward compatibility
type LegacyClient struct {
    client RegistryClient
}

// SearchPackages converts to old map-based format
func (l *LegacyClient) SearchPackages(query, tag, target string, limit int) ([]map[string]interface{}, error) {
    ctx, cancel := WithTimeout(context.Background())
    defer cancel()
    
    opts := SearchOptions{
        Query:  query,
        Tag:    tag,
        Target: target,
        Limit:  limit,
    }
    
    packages, err := l.client.SearchPackages(ctx, opts)
    if err != nil {
        return nil, err
    }
    
    results := make([]map[string]interface{}, len(packages))
    for i, pkg := range packages {
        results[i] = PackageToMap(&pkg)
    }
    
    return results, nil
}

// ... similar conversions for other methods
```

## Testing Requirements

### Unit Tests
1. Test interface implementation compliance
2. Test context cancellation handling
3. Test error type conversions
4. Test data structure conversions

### Integration Tests
1. Test with actual HTTP registry
2. Test authentication flow
3. Test timeout behavior
4. Test backward compatibility layer

### Regression Tests
1. All existing HTTP client tests should pass
2. CLI commands should work unchanged
3. Authentication should continue to work

### Cucumber Test Amendments

This phase requires ensuring all existing HTTP registry tests continue to pass with the refactored client.

**File**: `features/http-registry-operations.feature` (update existing)

Add scenarios to verify refactored HTTP client behavior:
```gherkin
Feature: HTTP Registry Operations
  HTTP registry operations should continue working after refactoring

  Background:
    Given I have a clean test environment
    And a running HTTP registry at "http://localhost:8080"

  Scenario: Search works with refactored HTTP client
    Given a registry "local" with URL "http://localhost:8080" and type "remote-http"
    And I use registry "local"
    When I run "rfh search test"
    Then the command should succeed

  Scenario: Publish works with refactored HTTP client
    Given a registry "local" with URL "http://localhost:8080" and type "remote-http"
    And I use registry "local"
    And I am authenticated to the registry
    And I have a package manifest
    When I run "rfh publish"
    Then the command should succeed

  Scenario: Context timeout is respected
    Given a registry "slow" with URL "http://localhost:8080" and type "remote-http"
    And the registry responds slowly
    When I run "rfh search test --timeout 1s"
    Then the command should fail with a timeout error

  Scenario: Error types are properly converted
    Given a registry "local" with URL "http://localhost:8080" and type "remote-http"
    When I run "rfh get nonexistent-package"
    Then the command should fail
    And the error should be "package not found"
```

**File**: `features/step_definitions/http_client_steps.js`

Add new step definitions:
```javascript
Given('the registry responds slowly', async function () {
  // Configure mock registry to add delay
  await this.mockRegistry.setResponseDelay(5000);
});

Then('the command should fail with a timeout error', async function () {
  assert(this.lastResult.error, 'Command should have failed');
  assert(this.lastResult.error.includes('timeout') || 
         this.lastResult.error.includes('context deadline exceeded'),
    `Expected timeout error, got: ${this.lastResult.error}`);
});

Then('the error should be {string}', async function (expectedError) {
  assert(this.lastResult.error, 'Command should have failed');
  assert(this.lastResult.error.toLowerCase().includes(expectedError.toLowerCase()),
    `Expected error containing "${expectedError}", got: ${this.lastResult.error}`);
});
```

**File**: `features/backward-compatibility.feature` (add to existing)

```gherkin
  Scenario: Legacy client code continues to work
    Given a registry "legacy" with URL "http://localhost:8080"
    # This registry has no type field, testing backward compatibility
    When I run "rfh search package"
    Then the command should succeed
    
  Scenario: Map-based responses are converted correctly
    Given a registry "local" with URL "http://localhost:8080" and type "remote-http"
    When I run "rfh search test --format json"
    Then the JSON output should be valid
    And the JSON should contain package objects with expected fields
```

## Success Criteria
- HTTPClient fully implements RegistryClient interface
- All CLI commands work with new interface
- Context support enables proper timeout/cancellation
- Backward compatibility maintained for existing code
- Error handling uses new error types consistently

## Dependencies
- Phase 1: Registry Type Core Architecture
- Phase 2: Registry Client Interface

## Migration Steps
1. Copy existing client.go to http.go
2. Update imports in all CLI commands
3. Run all tests to ensure nothing breaks
4. Remove old client.go once migration complete

## Risks
- **Risk**: Breaking existing CLI functionality
  **Mitigation**: Extensive testing, backward compatibility layer
  
- **Risk**: Performance regression
  **Mitigation**: Benchmark before/after, optimize hot paths

## Next Phase
Phase 4: Git Client Basic Operations - Implement core Git functionality for registry operations