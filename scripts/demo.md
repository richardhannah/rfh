# RuleStack POC Demo

This demonstrates the working RuleStack (rfh) package manager for AI rulesets.

## Prerequisites

1. **Database setup**: Either Docker Compose OR local Postgres
2. **Go installed**: For building the API and CLI

## Option 1: Docker Setup (Recommended)

```bash
# Start the complete development environment
./scripts/dev-up.sh

# The above will start:
# - PostgreSQL database with migrations
# - API server with hot reload (http://localhost:8080)
```

## Option 2: Local Database Setup

```bash
# Set up local Postgres database
createdb rulestack_dev

# Run migrations with Flyway
flyway -url=jdbc:postgresql://localhost:5432/rulestack_dev -user=postgres -password=yourpassword migrate

# Start API server
go run ./cmd/api
```

## Demo Steps

### 1. Build CLI
```bash
go build -buildvcs=false -o rfh ./cmd/cli
```

### 2. Set up development token
```bash
go run ./scripts/setup-dev.go
```
This creates a development token: `dev-token-12345`

### 3. Configure CLI registry
```bash
./rfh registry add local http://localhost:8080 dev-token-12345
./rfh registry list
```

### 4. Create a test package
```bash
./rfh init
```
This creates:
- `rulestack.json` (manifest)
- `rules/example-rule.md` (sample rule)
- `README.md` (documentation)

### 5. Pack the package
```bash
./rfh pack
```
Creates: `acme-example-rules-0.1.0.tgz`

### 6. Publish to registry
```bash
./rfh publish
```
Uploads the package to the registry.

### 7. Search for packages
```bash
./rfh search example
./rfh search "*" --target cursor
```

### 8. Test API directly
```bash
# Health check
curl http://localhost:8080/v1/health

# Search packages  
curl "http://localhost:8080/v1/packages?q=example"
```

## What You've Built

✅ **Complete package manager** for AI rulesets  
✅ **REST API** with authentication, search, and blob storage  
✅ **CLI tool** with registry management, publishing, and discovery  
✅ **Docker development environment** with hot reload  
✅ **Database schema** with proper relationships and indexing  
✅ **Archive handling** with SHA256 verification  
✅ **Multi-registry support** for public and private registries  

## Architecture

- **API Server**: Go + Gorilla/Mux + PostgreSQL + sqlx
- **CLI**: Go + Cobra + TOML config
- **Database**: PostgreSQL with Flyway migrations  
- **Storage**: Local filesystem (easily extensible to S3)
- **Auth**: Bearer tokens with SHA256 hashing
- **Packaging**: tar.gz archives with glob pattern support

## Production Ready Features

- Health checks and proper error handling
- Structured logging and CORS support  
- SHA256 integrity verification
- Scoped package names (@org/package)
- Multi-target support (cursor, claude-code, windsurf)
- Configurable registries and authentication
- Hot reloading development environment

## Next Steps for Production

1. **Deploy to Render** with PostgreSQL addon
2. **Add web UI** for package discovery
3. **Implement proper version resolution** (latest, semver ranges)
4. **Add package signing** with cosign/sigstore
5. **Add transparency log** for audit trail
6. **Build editor integrations** (VS Code extensions, etc.)

This POC demonstrates all core package manager functionality and is ready for stakeholder review!