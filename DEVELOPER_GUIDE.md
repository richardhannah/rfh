# RuleStack Developer Guide

This guide helps you get started developing RuleStack quickly and effectively.

## ⚡ **Ultra Quick Start**

**Want to start developing right now?**

```bash
# 1. Clone and enter directory
git clone <repo-url>
cd rulestack

# 2. Start everything
docker-compose up -d

# 3. Test that it works
powershell -File test-api-cli.ps1
# OR on Linux/Mac: bash test-api-cli.sh

# 4. Success? You're ready to develop! 🎉
```

**If the test script passes, your entire development environment is working correctly.**

## 🧪 **Testing Philosophy**

RuleStack strongly emphasizes **behavior testing** - tests that validate the system as a real user would experience it.

### What We Love ❤️

```bash
# End-to-end tests that start from scratch
docker-compose up -d
powershell -File test-api-cli.ps1

# Tests that validate complete user workflows
test-security-validation.ps1
test-new-feature.ps1
test-edge-cases.ps1
```

**These tests are valuable because they:**
- Catch integration bugs that unit tests miss
- Validate the actual user experience
- Test security, performance, and reliability together
- Ensure the system works in real conditions

### What We Like 👍

```bash
# Unit tests for critical security and algorithms
go test ./internal/security -v
go test ./internal/manifest -v
```

### What's Okay 😐

- Mocking and isolated component testing
- Tests that require extensive setup/teardown
- Tests that don't reflect real usage patterns

## 🛠 **Development Workflow**

### 1. Make Changes
Edit code, add features, fix bugs - whatever you're working on.

### 2. Test Your Changes
```bash
# Quick validation
go build ./cmd/cli
go build ./cmd/api

# Full system test
powershell -File test-api-cli.ps1
```

### 3. Add Behavior Tests
If you added a feature, create a test script that validates it:

```powershell
# test-my-feature.ps1
Write-Host "Testing my amazing feature..."

# Start with clean environment
docker-compose up -d
Start-Sleep 5

# Test the feature as a user would
./rfh init
./rfh my-new-command --option value

# Validate results
if (Test-Path "expected-result.txt") {
    Write-Host "✅ Feature works!"
} else {
    Write-Host "❌ Feature failed!"
    exit 1
}

# Cleanup
docker-compose down
```

### 4. Submit PR
- Your behavior test script
- Clear description of what the feature does
- How to test it manually

## 🏗 **Architecture Overview**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Client    │───▶│   API Server    │───▶│   PostgreSQL    │
│   (rfh)         │    │   (Go/Mux)      │    │   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       ▼                       │
         │              ┌─────────────────┐               │
         │              │  File Storage   │               │
         │              │  (.tgz files)   │               │
         │              └─────────────────┘               │
         │                                                │
         ▼                                                │
┌─────────────────┐                                       │
│  Local Project  │                                       │
│  (.rulestack/)  │◀──────────────────────────────────────┘
│  (CLAUDE.md)    │
└─────────────────┘
```

### Key Components

- **CLI (`cmd/cli/`)**: User interface, handles all user commands
- **API (`cmd/api/`)**: HTTP server, manages packages and storage
- **Security (`internal/security/`)**: Validates all packages for safety
- **Database (`internal/db/`)**: PostgreSQL storage layer
- **Client (`internal/client/`)**: HTTP client for CLI->API communication

## 🔒 **Security Development**

Security is built into every layer:

```bash
# Test security validation
go test ./internal/security -v

# Test with malicious packages
powershell -File test-security-simple.ps1
```

**Security features:**
- Package content validation (bluemonday for markdown)
- Path traversal prevention (zip slip protection)
- Executable detection and blocking
- File type allowlisting
- Size limits and encoding validation

When adding features, consider security implications and add behavior tests that validate security scenarios.

## 🐛 **Debugging**

### API Issues
```bash
# Check API logs
docker-compose logs api

# Check database
docker-compose exec postgres psql -U rulestack_user -d rulestack_dev
```

### CLI Issues
```bash
# Verbose output
./rfh --verbose command

# Check configuration
./rfh registry list
cat ~/.rfh/config.toml
```

### Test Issues
```bash
# Run individual test components
go test ./internal/security -v
go test ./internal/api -v

# Check Docker status
docker-compose ps
docker-compose logs
```

## 📝 **Adding New Features**

### 1. Understand the User Need
- What problem does this solve?
- How will users actually use this?
- What's the simplest implementation that works?

### 2. Design the Behavior Test First
Write the test script that validates your feature working end-to-end:

```bash
# This helps you think through the user experience
# before you write any implementation code
```

### 3. Implement
- Start with the simplest implementation that makes the test pass
- Add error handling and edge cases
- Consider security implications

### 4. Validate
```bash
# Your behavior test should pass
powershell -File test-my-feature.ps1

# Existing tests should still pass  
powershell -File test-api-cli.ps1
go test ./...
```

## 🚀 **Contributing Guidelines**

### We Want Your Contributions!

This project embraces **vibe coding**:
- ✅ **Working software** over perfect code
- ✅ **Real user value** over architectural purity  
- ✅ **Behavior tests** over extensive mocking
- ✅ **Practical solutions** over theoretical ideals

### Contribution Checklist

- [ ] Feature works (validated by behavior test)
- [ ] Existing tests still pass (`test-api-cli.ps1`)
- [ ] Code builds (`go build ./cmd/cli && go build ./cmd/api`)
- [ ] Clear description of what the feature does
- [ ] Behavior test script included for new features

**We'll help with code style and optimization during review. Focus on making it work first!**

## 🛠 **Useful Commands**

```bash
# Start development environment
docker-compose up -d

# Rebuild CLI after changes
go build -o dist/rfh.exe cmd/cli/main.go

# Full system test
powershell -File test-api-cli.ps1

# Security validation test
powershell -File test-security-simple.ps1

# Unit tests
go test ./...

# Clean restart
docker-compose down && docker-compose up -d

# Check logs
docker-compose logs -f api
```

## 🎯 **Next Steps**

1. **Run the quickstart** - make sure everything works
2. **Explore the codebase** - especially `cmd/cli/` and `cmd/api/`
3. **Run the test scripts** - understand what they validate
4. **Pick an issue or feature** - start with something small
5. **Write a behavior test** - think through the user experience
6. **Implement and submit PR** - we'll help with any rough edges

**Happy coding! 🚀**