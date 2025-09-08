package client

import "time"

// GitRegistryIndex represents the root index.json file
type GitRegistryIndex struct {
	Version      string                    `json:"version"`
	UpdatedAt    time.Time                 `json:"updated_at"`
	PackageCount int                       `json:"package_count"`
	Packages     map[string]GitPackageEntry `json:"packages"`
}

// GitPackageEntry represents a package entry in the index
type GitPackageEntry struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latest      string    `json:"latest"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
}

// GitPackageMetadata represents the metadata.json file
type GitPackageMetadata struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Latest      string              `json:"latest"`
	Versions    []GitVersionSummary `json:"versions"`
	Tags        []string            `json:"tags,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// GitVersionSummary represents a version in package metadata
type GitVersionSummary struct {
	Version     string    `json:"version"`
	SHA256      string    `json:"sha256"`
	Size        int64     `json:"size"`
	PublishedAt time.Time `json:"published_at"`
}

// GitManifest represents a version's manifest.json
type GitManifest struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Dependencies map[string]string      `json:"dependencies,omitempty"`
	SHA256       string                 `json:"sha256"`
	Size         int64                  `json:"size"`
	PublishedAt  time.Time              `json:"published_at"`
	Publisher    string                 `json:"publisher"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}