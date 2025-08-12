# Task 5: Database Connection and Models (1 hour)

## Objective
Create database connection setup and data models using sqlx for interacting with Postgres.

## Prerequisites
- Tasks 1-4 completed
- Database migrations run successfully
- Config module working
- Docker development environment running

## Checklist

### 1. Create Database Connection (15 minutes)
Create `internal/db/connection.go`:
```go
package db

import (
    "database/sql"
    
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq" // postgres driver
)

// DB holds the database connection
type DB struct {
    *sqlx.DB
}

// Connect establishes a connection to the database
func Connect(databaseURL string) (*DB, error) {
    sqlxDB, err := sqlx.Connect("postgres", databaseURL)
    if err != nil {
        return nil, err
    }
    
    // Test the connection
    if err := sqlxDB.Ping(); err != nil {
        sqlxDB.Close()
        return nil, err
    }
    
    return &DB{sqlxDB}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
    return db.DB.Close()
}

// Health checks if the database connection is healthy
func (db *DB) Health() error {
    return db.Ping()
}
```

- [ ] Create connection.go file
- [ ] Test that imports work correctly

### 2. Create Data Models (15 minutes)
Create `internal/db/models.go`:
```go
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
```

- [ ] Create models.go file
- [ ] Verify pq.StringArray import works for Postgres arrays

### 3. Create Package Repository (20 minutes)
Create `internal/db/packages.go`:
```go
package db

import (
    "database/sql"
    "fmt"
    "strings"
)

// GetOrCreatePackage gets existing package or creates new one
func (db *DB) GetOrCreatePackage(scope *string, name string) (*Package, error) {
    // First try to get existing
    pkg, err := db.GetPackage(scope, name)
    if err == nil {
        return pkg, nil
    }
    if err != sql.ErrNoRows {
        return nil, err
    }
    
    // Create new package
    query := `
        INSERT INTO packages (scope, name) 
        VALUES ($1, $2) 
        RETURNING id, scope, name, created_at`
        
    var newPkg Package
    err = db.Get(&newPkg, query, scope, name)
    if err != nil {
        return nil, err
    }
    
    return &newPkg, nil
}

// GetPackage retrieves a package by scope and name
func (db *DB) GetPackage(scope *string, name string) (*Package, error) {
    query := `SELECT id, scope, name, created_at FROM packages WHERE scope = $1 AND name = $2`
    
    var pkg Package
    err := db.Get(&pkg, query, scope, name)
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
func (db *DB) GetPackageVersion(scope *string, name string, version string) (*PackageVersion, error) {
    query := `
        SELECT pv.id, pv.package_id, pv.version, pv.description, pv.targets, pv.tags, 
               pv.sha256, pv.size_bytes, pv.blob_path, pv.created_at
        FROM package_versions pv
        JOIN packages p ON p.id = pv.package_id
        WHERE p.scope = $1 AND p.name = $2 AND pv.version = $3`
    
    var pkgVersion PackageVersion
    err := db.Get(&pkgVersion, query, scope, name, version)
    if err != nil {
        return nil, err
    }
    
    return &pkgVersion, nil
}

// SearchPackages searches for packages
func (db *DB) SearchPackages(query string, tag string, target string, limit int) ([]SearchResult, error) {
    sqlQuery := `
        SELECT DISTINCT p.id, p.scope, p.name, pv.version, pv.description, pv.targets, pv.tags, p.created_at
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
```

- [ ] Create packages.go file
- [ ] Review SQL queries for syntax errors

### 4. Create Token Repository (10 minutes)
Create `internal/db/tokens.go`:
```go
package db

import (
    "crypto/sha256"
    "fmt"
)

// CreateToken creates a new API token
func (db *DB) CreateToken(tokenHash string, name *string) (*Token, error) {
    query := `
        INSERT INTO tokens (token_hash, name) 
        VALUES ($1, $2) 
        RETURNING id, token_hash, name, created_at`
        
    var token Token
    err := db.Get(&token, query, tokenHash, name)
    if err != nil {
        return nil, err
    }
    
    return &token, nil
}

// ValidateToken checks if a token exists and is valid
func (db *DB) ValidateToken(tokenHash string) (*Token, error) {
    query := `SELECT id, token_hash, name, created_at FROM tokens WHERE token_hash = $1`
    
    var token Token
    err := db.Get(&token, query, tokenHash)
    if err != nil {
        return nil, err
    }
    
    return &token, nil
}

// HashToken creates a SHA256 hash of a token with salt
func HashToken(token string, salt string) string {
    h := sha256.New()
    h.Write([]byte(token + salt))
    return fmt.Sprintf("%x", h.Sum(nil))
}
```

- [ ] Create tokens.go file
- [ ] Verify crypto imports work

### 5. Create Integration Test (10 minutes)
Create `internal/db/db_test.go`:
```go
package db

import (
    "os"
    "testing"
)

func TestDatabaseConnection(t *testing.T) {
    dbURL := os.Getenv("TEST_DATABASE_URL")
    if dbURL == "" {
        t.Skip("TEST_DATABASE_URL not set, skipping integration test")
    }
    
    db, err := Connect(dbURL)
    if err != nil {
        t.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Test health check
    if err := db.Health(); err != nil {
        t.Fatalf("Database health check failed: %v", err)
    }
}

func TestPackageOperations(t *testing.T) {
    dbURL := os.Getenv("TEST_DATABASE_URL")
    if dbURL == "" {
        t.Skip("TEST_DATABASE_URL not set, skipping integration test")
    }
    
    db, err := Connect(dbURL)
    if err != nil {
        t.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Test package creation
    scope := "test"
    name := "test-package"
    
    pkg, err := db.GetOrCreatePackage(&scope, name)
    if err != nil {
        t.Fatalf("Failed to create package: %v", err)
    }
    
    if pkg.Name != name {
        t.Errorf("Expected package name %s, got %s", name, pkg.Name)
    }
    
    if pkg.Scope == nil || *pkg.Scope != scope {
        t.Errorf("Expected package scope %s, got %v", scope, pkg.Scope)
    }
    
    // Cleanup
    db.Exec("DELETE FROM packages WHERE id = $1", pkg.ID)
}
```

- [ ] Create test file
- [ ] Set TEST_DATABASE_URL if you want to run integration tests
- [ ] Run tests: `go test ./internal/db -v`

## Validation
Test the database layer:
```bash
# Run unit tests
go test ./internal/db -v

# Test connection manually
go run -c "
import (\"fmt\"; \"rulestack/internal/db\"; \"rulestack/internal/config\")
func main() {
    config.LoadEnvFile(\".env\")
    cfg := config.Load()
    db, err := db.Connect(cfg.DBURL)
    if err != nil { panic(err) }
    defer db.Close()
    fmt.Println(\"Database connected successfully\")
}"
```

## Acceptance Criteria
- [ ] Database connection established successfully
- [ ] All models defined with correct field mappings
- [ ] Package CRUD operations work
- [ ] Token validation functions work
- [ ] Search functionality implemented
- [ ] Integration tests pass (if TEST_DATABASE_URL set)
- [ ] No compilation errors
- [ ] Can ping database and verify health

## Time Estimate: ~60 minutes

## Next Task
Task 6: Basic API Handlers