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
func NewHTTPClient(baseURL, token string, verbose bool) *HTTPClient {
	baseURL = strings.TrimRight(baseURL, "/")

	return &HTTPClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
	}
}

// Type returns the registry type
func (c *HTTPClient) Type() config.RegistryType {
	return config.RegistryTypeHTTP
}

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
		Name:    getStringFromMap(result, "name"),
		Version: getStringFromMap(result, "version"),
		SHA256:  getStringFromMap(result, "sha256"),
		URL:     c.baseURL + "/v1/packages",
		Message: "Package published successfully",
	}, nil
}

// DownloadBlob downloads a blob by SHA256 hash
func (c *HTTPClient) DownloadBlob(ctx context.Context, sha256, destPath string) error {
	path := fmt.Sprintf("/v1/blobs/%s", sha256)

	resp, err := c.makeRequestWithContext(ctx, "GET", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return NewRegistryError(ErrNetworkError,
			fmt.Sprintf("download failed (status %d): %s", resp.StatusCode, string(body)))
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

// Health checks if the registry is healthy
func (c *HTTPClient) Health(ctx context.Context) error {
	resp, err := c.makeRequestWithContext(ctx, "GET", "/v1/health", nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewRegistryError(ErrNetworkError,
			fmt.Sprintf("registry health check failed (status %d)", resp.StatusCode))
	}

	return nil
}

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
	} else if c.verbose {
		fmt.Printf("⚠️  No token available - sending request without Authorization header\n")
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
		if resp.StatusCode >= 400 {
			// Log response headers for debugging auth issues
			authHeader := resp.Request.Header.Get("Authorization")
			if authHeader != "" {
				tokenPart := authHeader[7:] // Remove "Bearer "
				if len(tokenPart) > 20 {
					tokenPart = tokenPart[:20] + "..."
				}
				fmt.Printf("🔍 Request had Authorization: Bearer %s\n", tokenPart)
			} else {
				fmt.Printf("⚠️  Request had no Authorization header\n")
			}
		}
	}

	return resp, nil
}

// addFileToForm adds a file to a multipart form
func (c *HTTPClient) addFileToForm(writer *multipart.Writer, fieldName, filePath string) error {
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

// getStringFromMap safely gets a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}