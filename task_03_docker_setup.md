# Task 3: Docker Development Setup (1 hour)

## Objective
Set up Docker Compose for local development with PostgreSQL database and hot-reloading API server.

## Prerequisites
- Tasks 1-2 completed (environment setup and database migrations)
- Docker and Docker Compose installed
- Flyway migrations created

## Checklist

### 1. Create Dockerfile for API (15 minutes)
Create `Dockerfile`:
```dockerfile
# Use Go official image
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed for private repos and HTTPS)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o rulestack-api ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/rulestack-api .

# Create storage directory
RUN mkdir -p /root/storage

# Expose port
EXPOSE 8080

# Command to run
CMD ["./rulestack-api"]
```

- [ ] Create Dockerfile
- [ ] Verify Go version matches your local setup

### 2. Create Docker Compose for Development (20 minutes)
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  # PostgreSQL database
  postgres:
    image: postgres:15-alpine
    container_name: rulestack-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: rulestack_user
      POSTGRES_PASSWORD: rulestack_password
      POSTGRES_DB: rulestack_dev
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U rulestack_user -d rulestack_dev"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Flyway for migrations
  flyway:
    image: flyway/flyway:9-alpine
    container_name: rulestack-flyway
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./migrations:/flyway/sql
    command: >
      -url=jdbc:postgresql://postgres:5432/rulestack_dev
      -user=rulestack_user
      -password=rulestack_password
      migrate
    restart: "no"

  # RuleStack API (development mode with hot reload)
  api:
    build:
      context: .
      dockerfile: Dockerfile.dev
    container_name: rulestack-api
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      flyway:
        condition: service_completed_successfully
    ports:
      - "8080:8080"
    volumes:
      - .:/app
      - /app/tmp  # Exclude tmp directory from volume mount
    environment:
      DATABASE_URL: postgres://rulestack_user:rulestack_password@postgres:5432/rulestack_dev?sslmode=disable
      TOKEN_SALT: dev_salt_please_change_in_production
      STORAGE_PATH: /app/storage
      PORT: 8080
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
```

- [ ] Create docker-compose.yml
- [ ] Review environment variables

### 3. Create Development Dockerfile with Hot Reload (15 minutes)
Create `Dockerfile.dev`:
```dockerfile
FROM golang:1.21-alpine

# Install air for hot reloading and other development tools
RUN go install github.com/cosmtrek/air@latest

# Install wget for health checks
RUN apk add --no-cache wget

WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Create storage directory
RUN mkdir -p storage

# Expose port
EXPOSE 8080

# Use air for hot reloading in development
CMD ["air", "-c", ".air.toml"]
```

- [ ] Create Dockerfile.dev
- [ ] Verify air installation works

### 4. Create Air Configuration (10 minutes)
Create `.air.toml`:
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/api"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "storage", "migrations"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

- [ ] Create .air.toml
- [ ] Configure file watching patterns

### 5. Create Development Scripts (10 minutes)
Create `scripts/dev-up.sh`:
```bash
#!/bin/bash

echo "🚀 Starting RuleStack development environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Create necessary directories
mkdir -p storage
mkdir -p tmp

# Start services
echo "🐳 Starting Docker services..."
docker-compose up --build -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
echo "   - Postgres"
docker-compose exec -T postgres pg_isready -U rulestack_user -d rulestack_dev

echo "   - API server"
timeout=60
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if curl -s http://localhost:8080/v1/health > /dev/null; then
        break
    fi
    sleep 2
    elapsed=$((elapsed + 2))
done

if [ $elapsed -ge $timeout ]; then
    echo "❌ API server did not start within $timeout seconds"
    docker-compose logs api
    exit 1
fi

echo "✅ Development environment is ready!"
echo ""
echo "📋 Services:"
echo "   🐘 Postgres:  localhost:5432"
echo "   🌐 API:       http://localhost:8080"
echo ""
echo "📊 Useful commands:"
echo "   View logs:    docker-compose logs -f"
echo "   Stop:         docker-compose down"
echo "   Rebuild:      docker-compose up --build"
echo ""
echo "🔧 API will automatically reload when you change Go files!"

# Show development token
echo ""
echo "🔑 Development setup:"
echo "   Run: go run ./scripts/setup-dev.go"
echo "   Then: ./rfh registry add local http://localhost:8080 dev-token-12345"
```

Create `scripts/dev-down.sh`:
```bash
#!/bin/bash

echo "🛑 Stopping RuleStack development environment..."

docker-compose down

echo "✅ Development environment stopped."
echo ""
echo "💾 Data persisted in Docker volume 'rulestack_postgres_data'"
echo "   To remove all data: docker-compose down -v"
```

Make scripts executable:
```bash
chmod +x scripts/dev-up.sh scripts/dev-down.sh
```

- [ ] Create development scripts
- [ ] Make scripts executable
- [ ] Test scripts work correctly

### 6. Update .gitignore (5 minutes)
Add Docker-specific entries to `.gitignore`:
```gitignore
# Existing entries...

# Docker
tmp/
*.log

# Storage
storage/*
!storage/.gitkeep

# Air
.air.toml.bak

# Development
.env.local
```

Create `storage/.gitkeep`:
```bash
touch storage/.gitkeep
```

- [ ] Update .gitignore
- [ ] Create storage directory with .gitkeep

### 7. Create Docker Development Documentation (5 minutes)
Create `DOCKER.md`:
```markdown
# Docker Development Setup

This project uses Docker Compose for local development to provide consistent environments and easy setup.

## Prerequisites

- Docker Desktop or Docker Engine + Docker Compose
- Go 1.21+ (for CLI development)

## Quick Start

```bash
# Start development environment
./scripts/dev-up.sh

# In another terminal, set up development token
go run ./scripts/setup-dev.go

# Configure CLI to use local registry
go build -o rfh ./cmd/cli
./rfh registry add local http://localhost:8080 dev-token-12345

# Test the setup
./rfh search test
```

## Services

- **Postgres** (`localhost:5432`): Database with automatic migrations
- **API** (`localhost:8080`): Auto-reloading Go API server
- **Flyway**: Runs migrations on startup

## Development Workflow

1. Make changes to Go files in `cmd/api/` or `internal/`
2. Air automatically detects changes and rebuilds the API
3. API server restarts with new changes
4. Test your changes with the CLI or direct API calls

## Useful Commands

```bash
# View logs
docker-compose logs -f api
docker-compose logs -f postgres

# Access database directly
docker-compose exec postgres psql -U rulestack_user -d rulestack_dev

# Restart services
docker-compose restart api

# Rebuild and restart
docker-compose up --build -d api

# Stop everything
./scripts/dev-down.sh

# Stop and remove all data
docker-compose down -v
```

## Troubleshooting

### API won't start
```bash
# Check API logs
docker-compose logs api

# Check if database is ready
docker-compose exec postgres pg_isready -U rulestack_user -d rulestack_dev
```

### Port conflicts
If you have local Postgres running on 5432:
```bash
# Stop local postgres first, or change the port in docker-compose.yml
sudo service postgresql stop  # Ubuntu/Debian
brew services stop postgresql  # macOS
```

### Hot reload not working
```bash
# Restart the API service
docker-compose restart api

# Check air configuration
docker-compose exec api cat .air.toml
```
```

- [ ] Create Docker documentation
- [ ] Test all documented commands

## Validation
Test the complete Docker setup:
```bash
# Start development environment
./scripts/dev-up.sh

# Verify services are running
docker-compose ps
curl http://localhost:8080/v1/health

# Test hot reload by making a small change to cmd/api/main.go
# (add a log statement) and verify API restarts

# Test database connection
docker-compose exec postgres psql -U rulestack_user -d rulestack_dev -c "SELECT 1;"

# Stop environment
./scripts/dev-down.sh
```

## Acceptance Criteria
- [ ] Docker Compose starts all services successfully
- [ ] PostgreSQL is accessible and migrations run automatically
- [ ] API server starts and responds to health checks
- [ ] Hot reloading works when Go files are changed
- [ ] Scripts provide clear feedback and error handling
- [ ] Services restart properly after code changes
- [ ] Database data persists between container restarts
- [ ] All ports are correctly exposed
- [ ] Environment variables are properly configured
- [ ] Documentation covers common development scenarios

## Time Estimate: ~60 minutes

## Next Task
Task 4: Configuration Module (renumbered from previous Task 3)

## Notes
- This setup provides a production-like environment for development
- The API will automatically rebuild when you change Go source files
- Database schema is managed through Flyway migrations
- All data persists in Docker volumes between restarts
- Perfect for team development - everyone gets the same environment