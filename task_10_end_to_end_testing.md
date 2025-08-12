# Task 11: End-to-End Testing and Polish (1 hour)

## Objective
Complete the RuleStack POC by adding comprehensive testing, fixing integration issues, and adding polish for a production-ready demo.

## Prerequisites
- Tasks 1-10 completed
- All core functionality implemented
- API and CLI buildable

## Checklist

### 1. Create Test Token and Seed Data (15 minutes)
Create `scripts/setup-dev.go`:
```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    
    _ "github.com/lib/pq"
    
    "rulestack/internal/config"
    "rulestack/internal/db"
)

func main() {
    // Load environment
    config.LoadEnvFile(".env")
    cfg := config.Load()
    
    // Connect to database
    database, err := db.Connect(cfg.DBURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer database.Close()
    
    // Create a test token
    tokenValue := "dev-token-12345"
    tokenHash := db.HashToken(tokenValue, cfg.TokenSalt)
    
    // Check if token already exists
    var exists bool
    err = database.Get(&exists, "SELECT EXISTS(SELECT 1 FROM tokens WHERE token_hash = $1)", tokenHash)
    if err != nil {
        log.Fatal("Failed to check token:", err)
    }
    
    if !exists {
        // Insert test token
        name := "Development Token"
        _, err = database.CreateToken(tokenHash, &name)
        if err != nil {
            log.Fatal("Failed to create token:", err)
        }
        
        fmt.Printf("✅ Created development token\n")
        fmt.Printf("🔑 Token: %s\n", tokenValue)
        fmt.Printf("⚠️  Save this token - you'll need it for publishing\n\n")
    } else {
        fmt.Printf("✅ Development token already exists\n")
        fmt.Printf("🔑 Token: %s\n", tokenValue)
    }
    
    // Show setup instructions
    fmt.Printf("🚀 Setup complete! Next steps:\n")
    fmt.Printf("   1. Start API: go run ./cmd/api\n")
    fmt.Printf("   2. Add registry: ./rfh registry add local http://localhost:8080 %s\n", tokenValue)
    fmt.Printf("   3. Initialize package: ./rfh init\n")
    fmt.Printf("   4. Pack and publish: ./rfh pack && ./rfh publish\n")
}
```

Create script to run setup:
```bash
# scripts/setup.sh
#!/bin/bash
echo "🔧 Setting up RuleStack development environment..."

# Build setup tool
go build -o setup-dev ./scripts/setup-dev.go

# Run setup
./setup-dev

# Clean up
rm setup-dev

echo "✅ Development setup complete!"
```

- [ ] Create setup script
- [ ] Run setup to create test token
- [ ] Verify token is created in database

### 2. Create Integration Test Script (20 minutes)
Create `scripts/test-e2e.sh`:
```bash
#!/bin/bash
set -e

echo "🧪 Starting RuleStack End-to-End Test"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_PORT=8080
API_URL="http://localhost:$API_PORT"
TEST_TOKEN="dev-token-12345"
TEST_PACKAGE="@test/example-rules"

# Build binaries
echo "🔨 Building binaries..."
go build -o rfh-test ./cmd/cli
go build -o rulestack-api-test ./cmd/api

# Start API server in background
echo "🚀 Starting API server..."
export PORT=$API_PORT
./rulestack-api-test &
API_PID=$!

# Give server time to start
sleep 3

# Function to cleanup on exit
cleanup() {
    echo "🧹 Cleaning up..."
    kill $API_PID 2>/dev/null || true
    rm -f rfh-test rulestack-api-test
    rm -f test-*.tgz
    rm -f rfh.lock
    rm -rf .cursor .claude .windsurf
    rm -rf ~/.rfh/cache/@test
}

trap cleanup EXIT

# Test 1: Health check
echo -e "${YELLOW}Test 1: API Health Check${NC}"
if curl -s $API_URL/v1/health | grep -q "ok"; then
    echo -e "${GREEN}✅ API health check passed${NC}"
else
    echo -e "${RED}❌ API health check failed${NC}"
    exit 1
fi

# Test 2: Configure CLI
echo -e "${YELLOW}Test 2: CLI Configuration${NC}"
./rfh-test registry add local $API_URL $TEST_TOKEN
if ./rfh-test registry list | grep -q "local"; then
    echo -e "${GREEN}✅ Registry configuration passed${NC}"
else
    echo -e "${RED}❌ Registry configuration failed${NC}"
    exit 1
fi

# Test 3: Initialize package
echo -e "${YELLOW}Test 3: Package Initialization${NC}"
./rfh-test init
if [ -f "rulestack.json" ] && [ -d "rules" ]; then
    echo -e "${GREEN}✅ Package initialization passed${NC}"
else
    echo -e "${RED}❌ Package initialization failed${NC}"
    exit 1
fi

# Update manifest for testing
cat > rulestack.json << EOF
{
  "name": "$TEST_PACKAGE",
  "version": "1.0.0",
  "description": "Test ruleset for end-to-end testing",
  "targets": ["cursor", "claude-code"],
  "tags": ["test", "example"],
  "files": ["rules/**/*.md"],
  "license": "MIT"
}
EOF

# Test 4: Pack package
echo -e "${YELLOW}Test 4: Package Packing${NC}"
./rfh-test pack
if [ -f "test-example-rules-1.0.0.tgz" ]; then
    echo -e "${GREEN}✅ Package packing passed${NC}"
else
    echo -e "${RED}❌ Package packing failed${NC}"
    exit 1
fi

# Test 5: Publish package
echo -e "${YELLOW}Test 5: Package Publishing${NC}"
if ./rfh-test publish; then
    echo -e "${GREEN}✅ Package publishing passed${NC}"
else
    echo -e "${RED}❌ Package publishing failed${NC}"
    exit 1
fi

# Test 6: Search packages
echo -e "${YELLOW}Test 6: Package Search${NC}"
if ./rfh-test search "test" | grep -q "$TEST_PACKAGE"; then
    echo -e "${GREEN}✅ Package search passed${NC}"
else
    echo -e "${RED}❌ Package search failed${NC}"
    exit 1
fi

# Test 7: Add package
echo -e "${YELLOW}Test 7: Package Installation${NC}"
if ./rfh-test add "$TEST_PACKAGE@1.0.0"; then
    echo -e "${GREEN}✅ Package installation passed${NC}"
else
    echo -e "${RED}❌ Package installation failed${NC}"
    exit 1
fi

# Test 8: List installed packages
echo -e "${YELLOW}Test 8: List Installed Packages${NC}"
if ./rfh-test list | grep -q "$TEST_PACKAGE"; then
    echo -e "${GREEN}✅ Package listing passed${NC}"
else
    echo -e "${RED}❌ Package listing failed${NC}"
    exit 1
fi

# Test 9: Apply package to cursor
echo -e "${YELLOW}Test 9: Apply to Cursor${NC}"
if ./rfh-test apply "$TEST_PACKAGE" --target cursor; then
    if [ -d ".cursor/rules/$TEST_PACKAGE" ]; then
        echo -e "${GREEN}✅ Package application to Cursor passed${NC}"
    else
        echo -e "${RED}❌ Package application to Cursor failed (no files copied)${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Package application to Cursor failed${NC}"
    exit 1
fi

# Test 10: Apply package to claude-code
echo -e "${YELLOW}Test 10: Apply to Claude Code${NC}"
if ./rfh-test apply "$TEST_PACKAGE" --target claude-code; then
    if [ -d ".claude/rules/$TEST_PACKAGE" ]; then
        echo -e "${GREEN}✅ Package application to Claude Code passed${NC}"
    else
        echo -e "${RED}❌ Package application to Claude Code failed (no files copied)${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Package application to Claude Code failed${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 All tests passed! RuleStack is working correctly.${NC}"

# Show final status
echo ""
echo "📊 Test Summary:"
echo "   ✅ API Health Check"
echo "   ✅ CLI Configuration"
echo "   ✅ Package Initialization"
echo "   ✅ Package Packing"
echo "   ✅ Package Publishing"
echo "   ✅ Package Search"
echo "   ✅ Package Installation"
echo "   ✅ Package Listing"
echo "   ✅ Apply to Cursor"
echo "   ✅ Apply to Claude Code"
echo ""
echo "🚀 RuleStack POC is ready for demo!"
```

Make script executable:
```bash
chmod +x scripts/test-e2e.sh
```

- [ ] Create integration test script
- [ ] Make script executable
- [ ] Test script runs without errors

### 3. Fix Integration Issues (15 minutes)

#### Fix API Handler Issues
Update `internal/api/handlers.go` to add missing imports:
```go
// Add these imports at the top
import (
    "crypto/sha256"
    "path/filepath"
    "io"
    // ... existing imports
)
```

#### Fix CLI Add Command Issues
Update `internal/cli/add.go` to handle API response properly:
```go
// Update runAdd function to handle API response structure
func runAdd(packageSpec string) error {
    // ... existing code until API call ...
    
    // Get package information
    fmt.Printf("🔍 Resolving %s...\n", spec.FullName)
    packageInfo, err := c.GetPackage(spec.Scope, spec.Name)
    if err != nil {
        return fmt.Errorf("failed to get package info: %w", err)
    }
    
    // For now, use hardcoded version resolution
    // In a full implementation, this would resolve from package versions
    resolvedVersion := spec.Version
    if resolvedVersion == "" {
        resolvedVersion = "1.0.0" // Default for POC
    }
    
    // Get SHA256 from package info (this needs to be implemented in API response)
    sha256, ok := packageInfo["sha256"].(string)
    if !ok {
        // For POC, we'll need to implement proper version resolution
        // For now, return an error
        return fmt.Errorf("package version information not available. Please specify version with @1.0.0")
    }
    
    // ... rest of function remains the same
}
```

#### Create Missing Package Spec Parser
Add to `internal/cli/add.go` if not already present:
```go
// Add this if parsePackageSpec is referenced but not defined
func parsePackageSpec(spec string) (*PackageSpec, error) {
    // This function should already be in add.go from the previous task
    // If missing, copy from task 9 implementation
}
```

- [ ] Add missing imports to API handlers
- [ ] Fix CLI package resolution logic
- [ ] Test basic integration flow

### 4. Create Demo Script (10 minutes)
Create `scripts/demo.sh`:
```bash
#!/bin/bash

echo "🎬 RuleStack Demo - Registry for Humans"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

pause() {
    echo ""
    echo -e "${YELLOW}Press Enter to continue...${NC}"
    read -r
}

echo -e "${BLUE}What is RuleStack?${NC}"
echo "RuleStack (rfh) is a package manager for AI rulesets."
echo "It allows you to publish, share, and install AI rules for editors like Cursor and Claude Code."
pause

echo -e "${BLUE}Step 1: Check API Health${NC}"
echo "$ curl http://localhost:8080/v1/health"
curl -s http://localhost:8080/v1/health | jq
pause

echo -e "${BLUE}Step 2: Configure Registry${NC}"
echo "$ ./rfh registry add local http://localhost:8080 dev-token-12345"
./rfh registry add local http://localhost:8080 dev-token-12345
echo ""
echo "$ ./rfh registry list"
./rfh registry list
pause

echo -e "${BLUE}Step 3: Create a New Ruleset${NC}"
echo "$ ./rfh init"
./rfh init
echo ""
echo "Generated files:"
echo "- rulestack.json (manifest)"
echo "- rules/example-rule.md (sample rule)"
echo "- README.md (documentation)"
pause

echo -e "${BLUE}Step 4: Pack the Ruleset${NC}"
echo "$ ./rfh pack"
./rfh pack
echo ""
echo "Created archive ready for publishing"
pause

echo -e "${BLUE}Step 5: Publish to Registry${NC}"
echo "$ ./rfh publish"
./rfh publish
pause

echo -e "${BLUE}Step 6: Search for Rulesets${NC}"
echo "$ ./rfh search example"
./rfh search example
pause

echo -e "${BLUE}Step 7: Install a Ruleset${NC}"
echo "$ ./rfh add @acme/example-rules@0.1.0"
./rfh add @acme/example-rules@0.1.0
echo ""
echo "$ ./rfh list"
./rfh list
pause

echo -e "${BLUE}Step 8: Apply to Cursor${NC}"
echo "$ ./rfh apply @acme/example-rules --target cursor"
./rfh apply @acme/example-rules --target cursor
echo ""
echo "Rules are now available in .cursor/rules/"
ls -la .cursor/rules/@acme/example-rules/
pause

echo -e "${GREEN}🎉 Demo Complete!${NC}"
echo ""
echo "RuleStack Features Demonstrated:"
echo "✅ Package publishing and discovery"
echo "✅ Version management"
echo "✅ Multiple editor support"
echo "✅ Private and public registries"
echo "✅ Secure package distribution"
echo ""
echo "Ready for production deployment on Render!"
```

Make script executable:
```bash
chmod +x scripts/demo.sh
```

- [ ] Create demo script
- [ ] Test demo flow
- [ ] Ensure all commands work in sequence

## Validation
Run the complete test suite:
```bash
# Set up development environment
chmod +x scripts/setup.sh
./scripts/setup.sh

# Run end-to-end tests
./scripts/test-e2e.sh

# Run demo (with API server running)
go run ./cmd/api &
./scripts/demo.sh
```

## Final Polish Checklist
- [ ] All commands have helpful error messages
- [ ] Verbose mode provides useful debugging information  
- [ ] CLI help text is clear and comprehensive
- [ ] API returns proper HTTP status codes
- [ ] File permissions are set correctly (0o755 for dirs, 0o644 for files)
- [ ] Temp files and directories are cleaned up
- [ ] Configuration files are stored securely
- [ ] Package names and versions are validated
- [ ] SHA256 verification works correctly
- [ ] Lockfile maintains consistency

## Documentation Updates
Create or update these files:
- [ ] `README.md` - Overview and quick start
- [ ] `DEPLOYMENT.md` - Render deployment instructions  
- [ ] `API.md` - API documentation
- [ ] `CLI.md` - CLI command reference

## Acceptance Criteria
- [ ] End-to-end test script passes completely
- [ ] Demo script runs without errors
- [ ] All core functionality works: init, pack, publish, search, add, apply, list
- [ ] Registry management works properly
- [ ] Package installation and application works for multiple editors
- [ ] API handles authentication and file uploads correctly
- [ ] Error handling provides actionable feedback
- [ ] Performance is acceptable for POC (< 5 seconds for most operations)
- [ ] Security measures prevent directory traversal and other basic attacks

## Time Estimate: ~60 minutes

## Next Steps (Post-POC)
After completing this task, the RuleStack POC will be ready for:
1. Deployment to Render with Postgres
2. Adding real authentication tokens
3. Implementing proper version resolution
4. Adding signature verification
5. Building a web UI for package discovery
6. Adding more sophisticated rule validation

The POC demonstrates all core package manager functionality and is ready for stakeholder review and feedback.