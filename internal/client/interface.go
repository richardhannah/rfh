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