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
	PublishedAt  time.Time              `json:"published_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PublishResult contains information about a published package
type PublishResult struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
	URL     string `json:"url,omitempty"`    // For HTTP registries
	PRUrl   string `json:"pr_url,omitempty"` // For Git registries
	Message string `json:"message"`
}
