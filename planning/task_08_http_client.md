# Task 9: HTTP Client and Publishing (1 hour)

## Objective
Create HTTP client for communicating with the registry API and implement the publish command functionality.

## Prerequisites
- Tasks 1-8 completed
- CLI foundation working
- API server can be started and accepts requests

## Checklist

### 1. Create HTTP Client Package (20 minutes)
Create `internal/client/client.go`:
```go
package client

import (
    "bytes"
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
)

// Client represents an HTTP client for the RuleStack registry
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
    verbose    bool
}

// NewClient creates a new registry client
func NewClient(baseURL, token string) *Client {
    // Clean up base URL
    baseURL = strings.TrimRight(baseURL, "/")
    
    return &Client{
        baseURL: baseURL,
        token:   token,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        verbose: false,
    }
}

// SetVerbose enables verbose logging
func (c *Client) SetVerbose(verbose bool) {
    c.verbose = verbose
}

// makeRequest makes an HTTP request with authentication
func (c *Client) makeRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
    url := c.baseURL + path
    
    if c.verbose {
        fmt.Printf("🌐 %s %s\n", method, url)
    }
    
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add authentication header
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }
    
    // Set content type if provided
    if contentType != "" {
        req.Header.Set("Content-Type", contentType)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    
    return resp, nil
}

// SearchPackages searches for packages in the registry
func (c *Client) SearchPackages(query, tag, target string, limit int) ([]map[string]interface{}, error) {
    path := "/v1/packages"
    
    // Build query parameters
    params := url.Values{}
    if query != "" {
        params.Add("q", query)
    }
    if tag != "" {
        params.Add("tag", tag)
    }
    if target != "" {
        params.Add("target", target)
    }
    if limit > 0 {
        params.Add("limit", strconv.Itoa(limit))
    }
    
    if len(params) > 0 {
        path += "?" + params.Encode()
    }
    
    resp, err := c.makeRequest("GET", path, nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("search failed (status %d): %s", resp.StatusCode, string(body))
    }
    
    var results []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return results, nil
}

// GetPackage gets information about a specific package
func (c *Client) GetPackage(scope, name string) (map[string]interface{}, error) {
    var path string
    if scope != "" {
        path = fmt.Sprintf("/v1/packages/@%s/%s", scope, name)
    } else {
        path = fmt.Sprintf("/v1/packages/%s", name)
    }
    
    resp, err := c.makeRequest("GET", path, nil, "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("package not found")
    }
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(body))
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return result, nil
}

// PublishPackage publishes a package to the registry
func (c *Client) PublishPackage(manifestPath, archivePath string) (map[string]interface{}, error) {
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
    resp, err := c.makeRequest("POST", "/v1/packages", &buf, writer.FormDataContentType())
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("publish failed (status %d): %s", resp.StatusCode, string(body))
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return result, nil
}

// DownloadBlob downloads a blob by SHA256 hash
func (c *Client) DownloadBlob(sha256, destPath string) error {
    path := fmt.Sprintf("/v1/blobs/%s", sha256)
    
    resp, err := c.makeRequest("GET", path, nil, "")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(body))
    }
    
    // Create destination file
    outFile, err := os.Create(destPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer outFile.Close()
    
    // Copy data
    _, err = io.Copy(outFile, resp.Body)
    if err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    if c.verbose {
        fmt.Printf("📥 Downloaded %s\n", destPath)
    }
    
    return nil
}

// addFileToForm adds a file to a multipart form
func (c *Client) addFileToForm(writer *multipart.Writer, fieldName, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
    if err != nil {
        return err
    }
    
    _, err = io.Copy(part, file)
    return err
}

// Health checks if the registry is healthy
func (c *Client) Health() error {
    resp, err := c.makeRequest("GET", "/v1/health", nil, "")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("registry health check failed (status %d)", resp.StatusCode)
    }
    
    return nil
}
```

- [ ] Create client.go file
- [ ] Verify multipart form handling

### 2. Create Publish Command (20 minutes)
Create `internal/cli/publish.go`:
```go
package cli

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/client"
    "rulestack/internal/config"
    "rulestack/internal/manifest"
)

var (
    archivePath string
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
    Use:   "publish",
    Short: "Publish a ruleset to the registry",
    Long: `Publish a ruleset package to the configured registry.

This command will:
1. Read the rulestack.json manifest
2. Use the specified archive (or create one if not specified)
3. Upload both files to the registry
4. Validate the upload was successful

Requires authentication token to be configured in the registry.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return runPublish()
    },
}

func runPublish() error {
    // Load manifest
    manifest, err := manifest.Load("rulestack.json")
    if err != nil {
        return fmt.Errorf("failed to load manifest: %w", err)
    }
    
    // Determine archive path
    archive := archivePath
    if archive == "" {
        // Generate default archive name
        safeName := manifest.GetPackageName()
        if scope := manifest.GetScope(); scope != "" {
            safeName = scope + "-" + safeName
        }
        archive = fmt.Sprintf("%s-%s.tgz", safeName, manifest.Version)
    }
    
    // Check if archive exists
    if _, err := os.Stat(archive); os.IsNotExist(err) {
        return fmt.Errorf("archive not found: %s. Run 'rfh pack' first or specify --archive", archive)
    }
    
    // Get registry configuration
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Determine which registry to use
    registryName := cfg.Current
    if registry != "" {
        registryName = registry
    }
    
    if registryName == "" {
        return fmt.Errorf("no registry configured. Use 'rfh registry add' to add a registry")
    }
    
    reg, exists := cfg.Registries[registryName]
    if !exists {
        return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", registryName)
    }
    
    // Use token from flag or config
    authToken := reg.Token
    if token != "" {
        authToken = token
    }
    
    if authToken == "" {
        return fmt.Errorf("no authentication token configured for registry '%s'", registryName)
    }
    
    if verbose {
        fmt.Printf("📦 Publishing %s v%s\n", manifest.Name, manifest.Version)
        fmt.Printf("🌐 Registry: %s (%s)\n", registryName, reg.URL)
        fmt.Printf("📄 Archive: %s\n", archive)
    }
    
    // Create client
    c := client.NewClient(reg.URL, authToken)
    c.SetVerbose(verbose)
    
    // Test registry connection
    if err := c.Health(); err != nil {
        return fmt.Errorf("registry health check failed: %w", err)
    }
    
    // Publish package
    fmt.Printf("🚀 Publishing to %s...\n", reg.URL)
    result, err := c.PublishPackage("rulestack.json", archive)
    if err != nil {
        return fmt.Errorf("publish failed: %w", err)
    }
    
    // Show success message
    fmt.Printf("✅ Successfully published %s\n", manifest.Name)
    if version, ok := result["version"].(string); ok {
        fmt.Printf("📌 Version: %s\n", version)
    }
    if sha, ok := result["sha256"].(string); ok {
        fmt.Printf("🔒 SHA256: %s\n", sha)
    }
    
    if verbose {
        fmt.Printf("📋 Response: %+v\n", result)
    }
    
    return nil
}

func init() {
    publishCmd.Flags().StringVarP(&archivePath, "archive", "a", "", "path to archive file (defaults to <name>-<version>.tgz)")
}
```

- [ ] Create publish.go file
- [ ] Test that it handles missing archives correctly

### 3. Create Search Command (15 minutes)
Create `internal/cli/search.go`:
```go
package cli

import (
    "fmt"
    "strings"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/client"
    "rulestack/internal/config"
)

var (
    searchTag    string
    searchTarget string
    searchLimit  int
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
    Use:   "search <query>",
    Short: "Search for rulesets in the registry",
    Long: `Search for rulesets in the configured registry.

You can filter results by tags and targets to find rulesets that match
your specific needs.

Examples:
  rfh search security
  rfh search "secure coding" --tag=javascript
  rfh search linting --target=cursor
  rfh search react --limit=10`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return runSearch(args[0])
    },
}

func runSearch(query string) error {
    // Get registry configuration
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Determine which registry to use
    registryName := cfg.Current
    if registry != "" {
        registryName = registry
    }
    
    if registryName == "" {
        return fmt.Errorf("no registry configured. Use 'rfh registry add' to add a registry")
    }
    
    reg, exists := cfg.Registries[registryName]
    if !exists {
        return fmt.Errorf("registry '%s' not found. Use 'rfh registry list' to see available registries", registryName)
    }
    
    if verbose {
        fmt.Printf("🔍 Searching for: %s\n", query)
        fmt.Printf("🌐 Registry: %s (%s)\n", registryName, reg.URL)
        if searchTag != "" {
            fmt.Printf("🏷️  Tag filter: %s\n", searchTag)
        }
        if searchTarget != "" {
            fmt.Printf("🎯 Target filter: %s\n", searchTarget)
        }
    }
    
    // Create client (no auth needed for search)
    c := client.NewClient(reg.URL, reg.Token)
    c.SetVerbose(verbose)
    
    // Search packages
    results, err := c.SearchPackages(query, searchTag, searchTarget, searchLimit)
    if err != nil {
        return fmt.Errorf("search failed: %w", err)
    }
    
    if len(results) == 0 {
        fmt.Printf("No rulesets found matching '%s'\n", query)
        if searchTag != "" || searchTarget != "" {
            fmt.Printf("Try removing filters or using different search terms.\n")
        }
        return nil
    }
    
    // Display results
    fmt.Printf("📋 Found %d ruleset(s):\n\n", len(results))
    
    for _, result := range results {
        name, _ := result["name"].(string)
        version, _ := result["version"].(string)
        description, _ := result["description"].(string)
        
        fmt.Printf("📦 %s@%s\n", name, version)
        
        if description != "" {
            fmt.Printf("   %s\n", description)
        }
        
        // Display targets
        if targets, ok := result["targets"].([]interface{}); ok && len(targets) > 0 {
            var targetStrs []string
            for _, t := range targets {
                if str, ok := t.(string); ok {
                    targetStrs = append(targetStrs, str)
                }
            }
            if len(targetStrs) > 0 {
                fmt.Printf("   🎯 Targets: %s\n", strings.Join(targetStrs, ", "))
            }
        }
        
        // Display tags
        if tags, ok := result["tags"].([]interface{}); ok && len(tags) > 0 {
            var tagStrs []string
            for _, t := range tags {
                if str, ok := t.(string); ok {
                    tagStrs = append(tagStrs, str)
                }
            }
            if len(tagStrs) > 0 {
                fmt.Printf("   🏷️  Tags: %s\n", strings.Join(tagStrs, ", "))
            }
        }
        
        fmt.Printf("\n")
    }
    
    fmt.Printf("💡 Install with: rfh add <package-name>@<version>\n")
    
    return nil
}

func init() {
    searchCmd.Flags().StringVar(&searchTag, "tag", "", "filter by tag")
    searchCmd.Flags().StringVar(&searchTarget, "target", "", "filter by target (cursor, claude-code, etc.)")
    searchCmd.Flags().IntVar(&searchLimit, "limit", 20, "limit number of results")
}
```

- [ ] Create search.go file
- [ ] Test search output formatting

### 4. Add List Command (5 minutes)
Create `internal/cli/list.go`:
```go
package cli

import (
    "fmt"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/pkg"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List installed rulesets",
    Long: `List all rulesets that have been installed in the current workspace.

This command reads the rfh.lock file to show which rulesets are currently
installed, their versions, and installation paths.`,
    Aliases: []string{"ls"},
    RunE: func(cmd *cobra.Command, args []string) error {
        return runList()
    },
}

func runList() error {
    lockfile, err := pkg.LoadLockfile("rfh.lock")
    if err != nil {
        return fmt.Errorf("failed to load lockfile: %w", err)
    }
    
    if len(lockfile.Packages) == 0 {
        fmt.Printf("No rulesets installed.\n")
        fmt.Printf("💡 Install a ruleset with: rfh add <package-name>\n")
        return nil
    }
    
    fmt.Printf("📦 Installed rulesets:\n\n")
    
    for name, entry := range lockfile.Packages {
        fmt.Printf("📦 %s@%s\n", name, entry.Version)
        
        if len(entry.Targets) > 0 {
            fmt.Printf("   🎯 Targets: %v\n", entry.Targets)
        }
        
        if entry.InstallPath != "" {
            fmt.Printf("   📁 Path: %s\n", entry.InstallPath)
        }
        
        if entry.Registry != "" {
            fmt.Printf("   🌐 Registry: %s\n", entry.Registry)
        }
        
        if verbose && entry.SHA256 != "" {
            fmt.Printf("   🔒 SHA256: %s\n", entry.SHA256)
        }
        
        fmt.Printf("\n")
    }
    
    if lockfile.Registry != "" {
        fmt.Printf("🌐 Default registry: %s\n", lockfile.Registry)
    }
    
    return nil
}
```

- [ ] Create list.go file
- [ ] Test with empty lockfile

## Validation
Test the HTTP client and commands:
```bash
# Build CLI
go build -o rfh ./cmd/cli

# Start API server (in separate terminal)
go run ./cmd/api

# Test commands
./rfh registry add local http://localhost:8080 test-token
./rfh search test
./rfh publish
./rfh list
```

## Acceptance Criteria
- [ ] HTTP client handles authentication properly
- [ ] Multipart form upload works for publishing
- [ ] Search command formats results nicely
- [ ] Publish command validates prerequisites
- [ ] Error messages are helpful and actionable
- [ ] Verbose mode provides useful debugging info
- [ ] List command shows installed packages
- [ ] Client handles network errors gracefully

## Time Estimate: ~60 minutes

## Next Task
Task 10: Package Download and Installation