package client

import (
	"context"
)

// LegacyClient wraps the new client interface for backward compatibility
// This allows existing code that expects the old API to continue working
type LegacyClient struct {
	client RegistryClient
}

// NewLegacyClient creates a backward compatibility wrapper around a RegistryClient
func NewLegacyClient(client RegistryClient) *LegacyClient {
	return &LegacyClient{client: client}
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

// GetPackage converts to old map-based format
func (l *LegacyClient) GetPackage(name string) (map[string]interface{}, error) {
	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	pkg, err := l.client.GetPackage(ctx, name)
	if err != nil {
		return nil, err
	}

	return PackageToMap(pkg), nil
}

// GetPackageVersion converts to old map-based format
func (l *LegacyClient) GetPackageVersion(name, version string) (map[string]interface{}, error) {
	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	pkgVersion, err := l.client.GetPackageVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}

	return PackageVersionToMap(pkgVersion), nil
}

// PublishPackage converts to old map-based format
func (l *LegacyClient) PublishPackage(manifestPath, archivePath string) (map[string]interface{}, error) {
	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	result, err := l.client.PublishPackage(ctx, manifestPath, archivePath)
	if err != nil {
		return nil, err
	}

	return PublishResultToMap(result), nil
}

// DownloadBlob maintains the same signature
func (l *LegacyClient) DownloadBlob(sha256, destPath string) error {
	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	return l.client.DownloadBlob(ctx, sha256, destPath)
}

// Health maintains the same signature
func (l *LegacyClient) Health() error {
	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	return l.client.Health(ctx)
}

// SetVerbose is deprecated but maintained for compatibility
func (l *LegacyClient) SetVerbose(verbose bool) {
	// This is a no-op for the new interface since verbose is set during construction
	// We could potentially store this and create a new client, but that would be complex
	// For now, we'll just ignore this call since verbose mode is set when creating the client
}