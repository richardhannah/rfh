# RuleStack (RFH) - Registry for Humans

[![CI](https://github.com/username/rulestack/actions/workflows/ci.yml/badge.svg)](https://github.com/username/rulestack/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/username/rulestack)](https://goreportcard.com/report/github.com/username/rulestack)
[![codecov](https://codecov.io/gh/username/rulestack/branch/master/graph/badge.svg)](https://codecov.io/gh/username/rulestack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/username/rulestack)](https://golang.org/)

A package manager for AI rulesets, making it easy to share and discover AI rules for code editors like Claude Code, Cursor, and Windsurf.

## 🚀 Quick Start

### With Docker (Database Only - Recommended)
```bash
# Start PostgreSQL and run migrations
docker-compose up -d postgres flyway

# Build CLI
go build -buildvcs=false -o rfh ./cmd/cli

# Start API server
export DATABASE_URL="postgres://rulestack_user:rulestack_password@localhost:5432/rulestack_dev?sslmode=disable"
export TOKEN_SALT="dev_salt_please_change_in_production"
export STORAGE_PATH="./storage"
export PORT="8080"
mkdir -p storage
go run ./cmd/api &

# Set up development token
go run ./scripts/setup-dev.go

# Configure registry
./rfh registry add local http://localhost:8080 dev-token-12345

# Create and publish a package
./rfh init
./rfh pack
./rfh publish

# Search for packages
./rfh search example
```

### Manual Setup (Full)
```bash
# Set up database
createdb rulestack_dev
flyway -url=jdbc:postgresql://localhost:5432/rulestack_dev -user=postgres migrate

# Set environment variables
export DATABASE_URL="postgres://postgres@localhost:5432/rulestack_dev?sslmode=disable"
export TOKEN_SALT="dev_salt_please_change_in_production"
export STORAGE_PATH="./storage"
export PORT="8080"

# Create storage directory
mkdir -p storage

# Start API
go run ./cmd/api &

# Follow steps above for CLI usage
```

## 📦 What is RuleStack?

RuleStack is a package manager specifically designed for AI rulesets - the configuration files and prompts that guide AI coding assistants. Just like npm for JavaScript or pip for Python, RuleStack allows you to:

- **Publish** rulesets to public or private registries
- **Discover** rulesets created by the community  
- **Install** rulesets into your development environment
- **Version** rulesets with semantic versioning
- **Share** best practices across teams and projects

## 🎯 Supported Editors

- **Cursor** - AI-powered code editor
- **Claude Code** - Anthropic's coding assistant
- **Windsurf** - AI development environment  
- **GitHub Copilot** - Microsoft's AI pair programmer

## 🛠 Architecture

- **API Server**: Go + Gorilla/Mux + PostgreSQL
- **CLI**: Go + Cobra with TOML configuration
- **Database**: PostgreSQL with Flyway migrations
- **Storage**: Filesystem with sanitized filenames (extensible to S3/cloud storage)
- **Auth**: Bearer tokens with SHA256 hashing
- **Packaging**: Compressed tar archives with integrity verification

## 📋 CLI Commands

### Registry Management
```bash
rfh registry add <name> <url> [token]   # Add a registry
rfh registry list                       # List configured registries  
rfh registry use <name>                 # Set active registry
```

### Package Development
```bash
rfh init                               # Initialize new ruleset project
rfh pack                               # Create distributable archive
rfh publish                            # Publish to registry
```

### Package Discovery & Installation  
```bash
rfh search <query> [--tag] [--target]  # Search for rulesets
rfh add <package>[@version]            # Download ruleset
rfh apply <package> --target <editor>  # Apply to editor
rfh list                               # List installed rulesets
```

## 📁 Project Structure

```
rulestack/
├── cmd/
│   ├── api/           # API server
│   └── cli/           # CLI tool
├── internal/
│   ├── api/           # HTTP handlers
│   ├── client/        # HTTP client
│   ├── config/        # Configuration
│   ├── db/            # Database layer
│   ├── manifest/      # Package manifests
│   └── pkg/           # Utilities
├── migrations/        # Database schema
├── scripts/           # Development scripts
└── storage/           # File storage
```

## 🔧 Development

### Prerequisites
- Go 1.24+ (project uses Go 1.24.4)
- Docker & Docker Compose (for database)
- PostgreSQL (if not using Docker)

### Setup
```bash
# Clone repository
git clone <repo-url>
cd rulestack

# Start database only (Docker method)
docker-compose up -d postgres flyway

# OR use full Docker (note: currently has Go version compatibility issues)
# ./scripts/dev-up.sh

# Build tools
go build -buildvcs=false -o rfh ./cmd/cli
go build -buildvcs=false -o rulestack-api ./cmd/api

# Start API manually
export DATABASE_URL="postgres://rulestack_user:rulestack_password@localhost:5432/rulestack_dev?sslmode=disable"
export TOKEN_SALT="dev_salt_please_change_in_production"
export STORAGE_PATH="./storage"
export PORT="8080"
mkdir -p storage
go run ./cmd/api &
```

### Testing
```bash
# Run unit tests
go test ./... -v

# Run unit tests with coverage
go test ./... -cover

# Run linting
golangci-lint run

# Run integration tests (requires running API server)
go run ./scripts/setup-dev.go
./rfh registry add local http://localhost:8080 dev-token-12345
./rfh init && ./rfh pack && ./rfh publish
./rfh search example
```

### CI/CD Pipeline

The project uses GitHub Actions for continuous integration and deployment:

#### **CI Workflow** (`.github/workflows/ci.yml`)
- **Triggers**: Push to `master`/`main` branch, Pull Requests
- **Jobs**:
  - **Test**: Runs unit tests with coverage reporting
  - **Build**: Compiles CLI and API binaries for Linux
  - **Lint**: Runs `golangci-lint` for code quality checks
- **Features**:
  - Go module caching for faster builds
  - Race condition detection in tests
  - Coverage upload to Codecov
  - Binary artifact uploads

#### **Release Workflow** (`.github/workflows/release.yml`)
- **Triggers**: Git tags matching `v*.*.*` (e.g., `v1.0.0`)
- **Cross-platform builds**: Linux, macOS, Windows (AMD64/ARM64)
- **Auto-generated release notes** from commit messages
- **GitHub Release creation** with all binaries

#### **Dependency Management**
- **Dependabot**: Weekly updates for Go modules and GitHub Actions
- **Security**: Automated vulnerability scanning

## 🌍 Deployment

### Render.com (Recommended)
1. Connect repository to Render
2. Add PostgreSQL addon  
3. Set environment variables:
   ```
   DATABASE_URL=<render-postgres-url>
   TOKEN_SALT=<random-secret>
   STORAGE_PATH=/opt/render/project/src/storage
   PORT=10000
   ```
4. Deploy API service

### Docker
```bash
# Build production image
docker build -t rulestack-api .

# Run with PostgreSQL
docker run -e DATABASE_URL=... -e TOKEN_SALT=... -p 8080:8080 rulestack-api
```

## 📊 API Endpoints

### Public
- `GET /v1/health` - Health check
- `GET /v1/packages` - Search packages
- `GET /v1/packages/{scope}/{name}` - Package details
- `GET /v1/blobs/{sha256}` - Download package

### Authenticated  
- `POST /v1/packages` - Publish package

## 🔒 Security

- **Authentication**: Bearer token with SHA256 hashing
- **Integrity**: SHA256 verification for all packages
- **Input validation**: Manifest validation and path sanitization
- **Directory traversal protection**: Safe archive extraction
- **Token storage**: Salted hashes in database

## 🚧 Roadmap

- [ ] **Package signing** with cosign/sigstore
- [ ] **Web UI** for package discovery
- [ ] **Version resolution** (latest, semver ranges)  
- [ ] **Editor integrations** (VS Code, JetBrains)
- [ ] **Transparency log** for audit trail
- [ ] **Advanced search** with faceted filtering
- [ ] **Usage analytics** and download metrics
- [ ] **Private registries** with SSO integration

## 📜 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Ensure** your code passes all checks:
   ```bash
   # Run tests
   go test ./...
   
   # Run linting
   golangci-lint run
   
   # Check formatting
   go fmt ./...
   ```
4. **Commit** your changes (`git commit -m 'Add amazing feature'`)
5. **Push** to your branch (`git push origin feature/amazing-feature`)
6. **Open** a Pull Request

### CI Requirements
All Pull Requests must pass:
- ✅ **Unit Tests**: All tests must pass with coverage reporting
- ✅ **Linting**: Code must pass `golangci-lint` checks
- ✅ **Build**: Both CLI and API must compile successfully
- ✅ **Formatting**: Code must be properly formatted with `gofmt`

The CI pipeline automatically runs these checks on every PR.

## 📞 Support

- **Issues**: Create GitHub issue
- **Discussions**: GitHub Discussions
- **Email**: [Contact information]

---

**RuleStack** - Making AI rulesets accessible and shareable for everyone. 🚀