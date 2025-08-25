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
func (c *Client) GetPackage(name string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v1/packages/%s", name)

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

// GetPackageVersion gets information about a specific package version
func (c *Client) GetPackageVersion(name, version string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v1/packages/%s/versions/%s", name, version)

	resp, err := c.makeRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package version not found")
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