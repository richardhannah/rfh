package db

import (
	"time"

	"github.com/lib/pq"
)

// Package represents a package in the registry
type Package struct {
	ID        int       `db:"id" json:"id"`
	Scope     *string   `db:"scope" json:"scope"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// PackageVersion represents a specific version of a package
type PackageVersion struct {
	ID          int            `db:"id" json:"id"`
	PackageID   int            `db:"package_id" json:"package_id"`
	Version     string         `db:"version" json:"version"`
	Description *string        `db:"description" json:"description"`
	Targets     pq.StringArray `db:"targets" json:"targets"`
	Tags        pq.StringArray `db:"tags" json:"tags"`
	SHA256      *string        `db:"sha256" json:"sha256"`
	SizeBytes   *int           `db:"size_bytes" json:"size_bytes"`
	BlobPath    *string        `db:"blob_path" json:"blob_path"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
}

// Token represents an API authentication token
type Token struct {
	ID        int       `db:"id" json:"id"`
	TokenHash string    `db:"token_hash" json:"token_hash"`
	Name      *string   `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// PackageInfo combines package and version info for API responses
type PackageInfo struct {
	Package
	Versions []PackageVersion `json:"versions"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID          int            `db:"id" json:"id"`
	Scope       *string        `db:"scope" json:"scope"`
	Name        string         `db:"name" json:"name"`
	Version     string         `db:"version" json:"version"`
	Description *string        `db:"description" json:"description"`
	Targets     pq.StringArray `db:"targets" json:"targets"`
	Tags        pq.StringArray `db:"tags" json:"tags"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
}

// FullPackageName returns the full package name with scope
func (p *Package) FullPackageName() string {
	if p.Scope != nil && *p.Scope != "" {
		return "@" + *p.Scope + "/" + p.Name
	}
	return p.Name
}