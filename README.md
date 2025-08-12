# RuleStack (RFH) - Registry for Humans

A package manager for AI rulesets, making it easy to share and discover AI rules for code editors like Claude Code, Cursor, and Windsurf.

## 🚀 Quick Start

### With Docker (Recommended)
```bash
# Start development environment
./scripts/dev-up.sh

# Set up development token
go run ./scripts/setup-dev.go

# Build CLI
go build -buildvcs=false -o rfh ./cmd/cli

# Configure registry
./rfh registry add local http://localhost:8080 dev-token-12345

# Create and publish a package
./rfh init
./rfh pack
./rfh publish

# Search for packages
./rfh search example
```

### Manual Setup
```bash
# Set up database
createdb rulestack_dev
flyway -url=jdbc:postgresql://localhost:5432/rulestack_dev -user=postgres migrate

# Start API
go run ./cmd/api

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
- **Storage**: Filesystem (extensible to S3/cloud storage)
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
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL (if not using Docker)

### Setup
```bash
# Clone repository
git clone <repo-url>
cd rulestack

# Start development environment  
./scripts/dev-up.sh

# Build tools
go build -buildvcs=false -o rfh ./cmd/cli
go build -buildvcs=false -o rulestack-api ./cmd/api
```

### Testing
```bash
# Run unit tests
go test ./...

# Run integration tests
go run ./scripts/setup-dev.go
./rfh registry add local http://localhost:8080 dev-token-12345
./rfh init && ./rfh pack && ./rfh publish
./rfh search example
```

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

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)  
5. Open Pull Request

## 📞 Support

- **Issues**: Create GitHub issue
- **Discussions**: GitHub Discussions
- **Email**: [Contact information]

---

**RuleStack** - Making AI rulesets accessible and shareable for everyone. 🚀