package db

import (
	"database/sql"
	"fmt"
)

// GetOrCreatePackage gets existing package or creates new one
func (db *DB) GetOrCreatePackage(name string) (*Package, error) {
	// First try to get existing
	pkg, err := db.GetPackage(name)
	if err == nil {
		return pkg, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new package
	query := `
        INSERT INTO packages (name) 
        VALUES ($1) 
        RETURNING id, name, created_at`

	var newPkg Package
	err = db.Get(&newPkg, query, name)
	if err != nil {
		return nil, err
	}

	return &newPkg, nil
}

// GetPackage retrieves a package by name
func (db *DB) GetPackage(name string) (*Package, error) {
	query := `SELECT id, name, created_at FROM packages WHERE name = $1`

	var pkg Package
	err := db.Get(&pkg, query, name)
	if err != nil {
		return nil, err
	}

	return &pkg, nil
}

// CreatePackageVersion creates a new package version
func (db *DB) CreatePackageVersion(version PackageVersion) (*PackageVersion, error) {
	query := `
        INSERT INTO package_versions 
        (package_id, version, description, targets, tags, sha256, size_bytes, blob_path)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, package_id, version, description, targets, tags, sha256, size_bytes, blob_path, created_at`

	var newVersion PackageVersion
	err := db.Get(&newVersion, query,
		version.PackageID,
		version.Version,
		version.Description,
		version.Targets,
		version.Tags,
		version.SHA256,
		version.SizeBytes,
		version.BlobPath,
	)

	if err != nil {
		return nil, err
	}

	return &newVersion, nil
}

// GetPackageVersion retrieves a specific version of a package
func (db *DB) GetPackageVersion(name string, version string) (*PackageVersion, error) {
	query := `
		SELECT pv.id, pv.package_id, pv.version, pv.description, pv.targets, pv.tags, 
			   pv.sha256, pv.size_bytes, pv.blob_path, pv.created_at
		FROM package_versions pv
		JOIN packages p ON p.id = pv.package_id
		WHERE p.name = $1 AND pv.version = $2`

	fmt.Printf("[DEBUG] GetPackageVersion SQL query: %s\n", query)
	fmt.Printf("[DEBUG] GetPackageVersion parameters: [%s, %s]\n", name, version)
	
	var pkgVersion PackageVersion
	err := db.Get(&pkgVersion, query, name, version)
	if err != nil {
		fmt.Printf("[ERROR] GetPackageVersion SQL error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[DEBUG] GetPackageVersion found: %+v\n", pkgVersion)
	return &pkgVersion, nil
}

// SearchPackages searches for packages
func (db *DB) SearchPackages(query string, tag string, target string, limit int) ([]SearchResult, error) {
	sqlQuery := `
        SELECT DISTINCT p.id, p.name, pv.version, pv.description, pv.targets, pv.tags, p.created_at
        FROM packages p
        JOIN package_versions pv ON p.id = pv.package_id
        WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	// Add search conditions
	if query != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND (p.name ILIKE $%d OR pv.description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+query+"%")
	}

	if tag != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND $%d = ANY(pv.tags)", argCount)
		args = append(args, tag)
	}

	if target != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND $%d = ANY(pv.targets)", argCount)
		args = append(args, target)
	}

	sqlQuery += " ORDER BY p.created_at DESC"

	if limit > 0 {
		argCount++
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	var results []SearchResult
	err := db.Select(&results, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	return results, nil
}