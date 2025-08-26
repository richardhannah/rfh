# RuleStack (RFH) - Registry for Humans

[![CI](https://github.com/username/rulestack/actions/workflows/ci.yml/badge.svg)](https://github.com/username/rulestack/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/username/rulestack)](https://goreportcard.com/report/github.com/username/rulestack)
[![codecov](https://codecov.io/gh/username/rulestack/branch/master/graph/badge.svg)](https://codecov.io/gh/username/rulestack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/username/rulestack)](https://golang.org/)

A secure package manager for AI rulesets, making it easy to share and discover AI rules for code editors like Claude Code, Cursor, and Windsurf. Built with enterprise-grade security validation and automatic Claude Code integration.

## 🚀 Quick Start

## 👨‍💻 Developer Quickstart

**TL;DR**: Want to start developing? Run this and you're good to go:

```bash
# 1. Start the development environment
docker-compose up -d

# 2. Run the comprehensive test suite
powershell -File run-tests.ps1
# OR on Linux/Mac: 
# bash run-tests.sh

# 3. If the test passes, you're ready to develop! 🎉
```

**What this does:**
- Starts PostgreSQL database with migrations
- Builds and runs the API server  
- Runs comprehensive BDD (Behavior Driven Development) tests including:
  - Package creation (rfh init), packing (rfh pack), and publishing (rfh publish)
  - Registry management (rfh registry add/list/use/remove)
  - User authentication (rfh auth login/register)
  - Complete end-to-end workflows with real CLI commands

**If the test script passes, your development environment is working correctly.**

📖 **For detailed development guidance, see [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)**

### Production Quick Start

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

RuleStack is a secure package manager specifically designed for AI rulesets - the configuration files and prompts that guide AI coding assistants. Just like npm for JavaScript or pip for Python, RuleStack allows you to:

- **Publish** rulesets to public or private registries
- **Discover** rulesets created by the community  
- **Install** rulesets with enterprise-grade security validation
- **Version** rulesets with semantic versioning
- **Share** best practices across teams and projects
- **Integrate** automatically with Claude Code via CLAUDE.md

## ✨ Key Features

### 🔒 **Enterprise Security**
- **Malware Protection**: Blocks executables, scripts, and malicious content
- **Path Traversal Prevention**: Protects against zip slip attacks
- **Content Sanitization**: Uses bluemonday to sanitize markdown (XSS protection)
- **File Type Validation**: Only allows safe file types (.md, .txt, .json)
- **Size Limits**: Prevents zip bombs and oversized packages

### 🤖 **AI Editor Integration**
- **Automatic CLAUDE.md Updates**: Seamlessly integrates with Claude Code
- **Core Rules System**: Ships with baseline rules for best practices
- **Versioned Rule References**: Links to actual files, not assumptions
- **Project-Specific Rules**: Supports both global and project rules

### 🚀 **Developer Experience**  
- **Git-like Workflow**: Familiar `init`, `pack`, `publish` commands
- **Explicit Project Roots**: Clear project boundaries like `git init`
- **Comprehensive Testing**: Full end-to-end validation
- **Zero-Config Security**: Security validation happens automatically

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
rfh registry remove <name>              # Remove a registry
```

### Package Development
```bash
rfh init [--force]                     # Initialize new ruleset project (like git init)
rfh pack                               # Create distributable archive  
rfh publish                            # Publish to registry
```

### Package Discovery & Installation  
```bash
rfh search <query> [--tag] [--target]  # Search for rulesets
rfh add <package>[@version]            # Download and install ruleset (auto-updates CLAUDE.md)
```

**New Workflow**: RuleStack now uses explicit project roots like Git:
1. **`rfh init`** - Creates project structure, CLAUDE.md, and core rules
2. **`rfh add <package>`** - Installs to `.rulestack/package.version/` and updates CLAUDE.md
3. **Built-in Security** - All packages validated automatically during installation

## 📁 Project Structure

### Repository Structure
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
│   ├── pkg/           # Package utilities
│   └── security/      # Security validation (NEW)
├── migrations/        # Database schema
├── scripts/           # Development scripts
├── storage/           # File storage
├── cucumber-testing/  # BDD test suite with Cucumber scenarios
├── run-tests.ps1      # BDD test runner (PowerShell)
├── run-tests.sh       # BDD test runner (Bash)
└── planning/          # Design documents
```

### RuleStack Project Structure (after `rfh init`)
```
my-rules/
├── rulestack.json              # Package manifest
├── CLAUDE.md                   # Claude Code integration (AUTO-GENERATED)
├── rules/
│   └── example-rule.md         # Your rule files
├── .rulestack/                 # Installed dependencies
│   ├── core.v1.0.0/           # Core rules (auto-installed)
│   │   └── core_rules.md
│   └── package.version/        # Installed packages
│       └── rules/
└── rulestack.lock.json        # Dependency lock file
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
# Run all unit tests
go test ./... -v

# Run security validation tests
go test ./internal/security -v

# Run unit tests with coverage  
go test ./... -cover

# Run linting
golangci-lint run

# Run comprehensive BDD test suite (recommended)
powershell -File run-tests.ps1
# OR: bash run-tests.sh
```

**BDD Test Coverage:**
- ✅ **52 scenarios** covering all CLI functionality
- ✅ **383 test steps** with real command execution  
- ✅ **Complete workflows**: init → pack → publish → registry management
- ✅ **Error handling**: Authentication, validation, network issues
- ✅ **Cross-platform**: Windows PowerShell and Unix Bash runners

**The test scripts validate:**
- Package creation, publishing, and installation
- Security validation (malware protection, path traversal, etc.)
- Claude Code integration (CLAUDE.md updates)
- Registry management
- Complete environment isolation and cleanup

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

RuleStack implements enterprise-grade security validation to protect against malicious packages:

### Package Security Validation
- **Path Traversal Protection**: Blocks `../../../etc/passwd` attacks (zip slip prevention)
- **Executable Detection**: Rejects scripts (.sh, .bat, .ps1), binaries (.exe, .dll), and executables (ELF, PE headers)
- **Content Sanitization**: Uses bluemonday to sanitize markdown and block XSS attacks
- **File Type Allowlist**: Only permits safe file types (.md, .txt, .json)
- **Size Limits**: Prevents zip bombs (1MB per file, 10MB total, max 100 files)
- **Encoding Validation**: Requires valid UTF-8, rejects NUL bytes
- **Symlink Protection**: Only allows regular files and directories

### Authentication & Integrity
- **Bearer Token Auth**: SHA256 hashing with salt
- **Package Integrity**: SHA256 verification for all packages  
- **Input Validation**: Comprehensive manifest and path sanitization
- **Token Storage**: Salted hashes in database
- **Secure Extraction**: Defense-in-depth package extraction

### Security Testing
RuleStack includes comprehensive security tests covering all attack vectors:
```bash
go test ./internal/security -v
```

**100% test coverage** for security validation including malicious markdown, path traversal, executables, and more.

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

We welcome and **strongly encourage** contributions! This project embraces the **vibe coding** philosophy - we care more about working software than perfect code.

### 🎯 **Our Philosophy**
- **Behavior over Beauty**: We prioritize working features over perfect code style
- **Vibe Coding Encouraged**: If it works and improves the project, we want it
- **BDD Testing Preferred**: We use Cucumber BDD tests that validate complete user workflows
- **Real-World Focus**: Tests should examine the system as users actually use it
- **Scenario-Based Testing**: Write tests as user stories with Given/When/Then scenarios

### 📝 **Contribution Guidelines**

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Build something that works** - don't worry about perfection
4. **Write BDD tests** for new features using Cucumber scenarios:
   ```bash
   # Run the full BDD test suite:
   powershell -File run-tests.ps1
   
   # Add new scenarios to existing feature files:
   # cucumber-testing/features/your-feature.feature
   
   # Example Cucumber scenario:
   # Scenario: User can create a new package
   #   Given I have a clean project directory
   #   When I run "rfh init --name my-package"
   #   Then I should see "Initialized RuleStack project"
   #   And a file "rulestack.json" should be created
   ```
5. **Ensure basic quality**:
   ```bash
   # Run existing tests
   go test ./...
   
   # Check that it builds
   go build ./cmd/cli
   go build ./cmd/api
   ```
6. **Commit** with a clear message (`git commit -m 'Add amazing feature'`)
7. **Push** to your branch (`git push origin feature/amazing-feature`)
8. **Open** a Pull Request

### 🧪 **Testing Philosophy**

**We strongly encourage behavior testing scripts** that validate complete workflows:

✅ **LOVE**: Tests that start from `docker-compose up` and validate entire user journeys
✅ **LOVE**: Scripts that test real integration scenarios  
✅ **LOVE**: End-to-end validation that covers security, UI, and business logic
✅ **LOVE**: Tests that would catch regressions a real user would experience

👍 **LIKE**: Unit tests for critical algorithms and security validation
😐 **OKAY**: Mocking and isolated component testing

### 🚀 **CI Requirements**
All Pull Requests must pass:
- ✅ **BDD Test Suite**: `run-tests.ps1` must pass (all 52 scenarios)
- ✅ **Build**: Both CLI and API must compile successfully  
- ✅ **Core Tests**: Security and critical unit tests must pass
- 📝 **New Feature Tests**: Include Cucumber scenarios for new features

**BDD Testing Guidelines:**
- Add scenarios to existing `.feature` files when extending functionality
- Create new `.feature` files for entirely new commands or workflows  
- Use Given/When/Then format with clear, user-focused language
- Test both success paths and error conditions

**We're more interested in working software than perfect lint scores.** If your code works and has good behavior test coverage, we'll help you with any style issues during review.

## 📞 Support

- **Issues**: Create GitHub issue
- **Discussions**: GitHub Discussions
- **Email**: [Contact information]

---

**RuleStack** - Making AI rulesets accessible and shareable for everyone. 🚀