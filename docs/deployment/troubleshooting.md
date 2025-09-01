# Troubleshooting Guide

This guide helps you resolve common issues with RFH.

## Common Issues

### Installation Issues

#### Binary Not Found / Command Not Found

**Error**: `rfh: command not found` or `'rfh' is not recognized`

**Solution**:
```bash
# Check if rfh is in PATH
which rfh  # Linux/Mac
where rfh  # Windows

# Add to PATH if needed
export PATH=$PATH:/path/to/rfh  # Linux/Mac
# Or add to system PATH on Windows

# Verify installation
rfh --version
```

#### Permission Denied

**Error**: `permission denied: ./rfh`

**Solution**:
```bash
# Make executable
chmod +x rfh

# Or run with explicit path
./rfh --version
```

#### SSL Certificate Issues

**Error**: `x509: certificate signed by unknown authority`

**Solutions**:
```bash
# Option 1: Update certificates (recommended)
# Linux: sudo apt-get update && sudo apt-get install ca-certificates
# Mac: brew install ca-certificates

# Option 2: Temporary workaround (testing only)
rfh registry add test https://registry.example.com --insecure

# Option 3: Set certificate file
export SSL_CERT_FILE=/path/to/cert.pem
```

### Configuration Issues

#### Config File Issues

**Error**: `failed to load config: permission denied`

**Solution**:
```bash
# Fix permissions
chmod 600 ~/.rfh/config.toml
chmod 700 ~/.rfh

# If file is corrupted, reset
mv ~/.rfh/config.toml ~/.rfh/config.toml.backup
rfh registry add default https://registry.example.com
```

#### TOML Parse Errors

**Error**: `toml: line 5: expected key separator '='`

**Solution**:
```bash
# Validate config file
cat ~/.rfh/config.toml

# Common issues:
# - Missing quotes around strings
# - Invalid characters
# - Incorrect indentation

# Reset if corrupted
rm ~/.rfh/config.toml
rfh registry add default https://registry.example.com
```

#### Registry Not Found

**Error**: `registry 'myregistry' not found`

**Solutions**:
```bash
# Check configured registries
rfh registry list

# Add missing registry
rfh registry add myregistry https://registry.example.com

# Set active registry
rfh registry use myregistry
```

### Authentication Issues

#### Login Failed

**Error**: `login failed: 401 Unauthorized`

**Solutions**:
```bash
# Check credentials
rfh auth login --username=correctuser

# Verify registry URL
rfh registry list

# Check registry connectivity
curl -I https://registry.example.com/v1/health

# Reset auth if needed
rfh auth logout
rfh auth login
```

#### Token Expired

**Error**: `authentication failed: token expired`

**Solution**:
```bash
# Re-authenticate
rfh auth logout
rfh auth login

# Check token status
rfh auth status
```

#### No Active Registry

**Error**: `no active registry configured`

**Solution**:
```bash
# Add and set registry
rfh registry add default https://registry.example.com
rfh registry use default

# Verify
rfh registry list
```

### Package Operations

#### Pack Command Issues

**Error**: `file must be a valid .mdc file`

**Solution**:
```bash
# Check file extension
ls -la *.mdc

# Rename file if needed
mv rules.txt rules.mdc

# Verify file content
file rules.mdc  # Should show UTF-8 text
```

**Error**: `failed to read input` (interactive mode)

**Solution**:
```bash
# Use non-interactive mode
rfh pack --file=rules.mdc --package=my-rules

# Or ensure stdin is available for interactive input
```

#### Version Validation Failed

**Error**: `version validation failed: new version 1.1.0 must be greater than current version 2.0.0`

**Solutions**:
```bash
# Check current version
rfh status

# Use higher version
rfh pack --file=rules.mdc --package=my-rules --version=2.1.0

# Or let RFH auto-increment
rfh pack --file=rules.mdc --package=my-rules
```

#### File Conflicts

**Error**: `file conflict.mdc already exists in package my-rules@1.0.0`

**Solutions**:
```bash
# Use different filename
mv conflict.mdc new-feature.mdc
rfh pack --file=new-feature.mdc --package=my-rules

# Or increment version to replace
rfh pack --file=conflict.mdc --package=my-rules --version=1.1.0
```

### Publishing Issues

#### No Staged Packages

**Error**: `No staged packages found`

**Solution**:
```bash
# Create packages first
rfh pack --file=rules.mdc --package=my-rules

# Check staged packages
rfh status

# Then publish
rfh publish
```

#### Publish Failed

**Error**: `publish failed: 413 Request Entity Too Large`

**Solution**:
```bash
# Check package size
ls -lh .rulestack/staged/*.tgz

# Split large packages or remove unnecessary files
# Package size limit is typically 10MB
```

#### Authentication Required

**Error**: `authentication required for publish`

**Solution**:
```bash
# Login first
rfh auth login

# Verify authentication
rfh auth status

# Then publish
rfh publish
```

### Network Issues

#### Connection Timeout

**Error**: `dial tcp: i/o timeout`

**Solutions**:
```bash
# Check network connectivity
ping registry.example.com

# Check registry status
curl -I https://registry.example.com/v1/health

# Try different registry
rfh registry use backup-registry

# Check firewall/proxy settings
```

#### DNS Resolution Failed

**Error**: `no such host: registry.example.com`

**Solutions**:
```bash
# Check DNS resolution
nslookup registry.example.com

# Try with IP address temporarily
rfh registry add temp http://192.168.1.100:8080

# Check /etc/hosts or DNS settings
```

### Development Issues

#### Test Failures

**Error**: Tests failing in development environment

**Solutions**:
```bash
# Rebuild binary
go build -o dist/rfh ./cmd/cli

# Clean and restart services
docker-compose down
docker-compose up -d --build

# Run specific test
go test ./internal/security -v

# Check test dependencies
cd cucumber-testing && npm install
```

#### Docker Issues

**Error**: `port 5432 already in use`

**Solutions**:
```bash
# Check what's using the port
netstat -tulpn | grep 5432

# Stop conflicting service
sudo service postgresql stop

# Or use different port
# Edit docker-compose.yml port mapping
```

**Error**: `permission denied while connecting to Docker daemon`

**Solution**:
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
# Log out and back in

# Or use sudo temporarily
sudo docker-compose up -d
```

## Debugging Tools

### Verbose Output

Enable detailed logging:

```bash
# Global verbose flag
rfh --verbose command

# Debug environment variable
RFH_DEBUG=1 rfh command

# Both together for maximum detail
RFH_DEBUG=1 rfh --verbose pack --file=test.mdc --package=debug
```

### Configuration Inspection

Check current configuration:

```bash
# List registries
rfh registry list

# Check auth status
rfh auth status

# Show effective config location
rfh --help | grep config

# Validate config file
cat ~/.rfh/config.toml
```

### Network Debugging

Test registry connectivity:

```bash
# Health check
curl -v https://registry.example.com/v1/health

# Test authentication
curl -H "Authorization: Bearer $TOKEN" https://registry.example.com/v1/packages

# Check TLS
openssl s_client -connect registry.example.com:443
```

### Package Debugging

Inspect package contents:

```bash
# List staged packages
ls -la .rulestack/staged/

# Extract and inspect archive
cd /tmp
tar -tzf path/to/package.tgz
tar -xzf path/to/package.tgz
```

## Getting Help

### Check Logs

RFH logs errors to stderr. Capture full output:

```bash
# Redirect both stdout and stderr
rfh command > output.log 2>&1

# View errors only
rfh command 2> errors.log
```

### System Information

When reporting issues, include:

```bash
# RFH version
rfh --version

# Operating system
uname -a  # Linux/Mac
systeminfo | findstr /B /C:"OS Name" /C:"OS Version"  # Windows

# Go version (for development)
go version

# Docker version (for development)
docker --version
docker-compose --version
```

### Create Minimal Reproduction

For complex issues:

```bash
# Create clean environment
mkdir rfh-debug && cd rfh-debug

# Minimal config
export RFH_CONFIG_PATH="$PWD/debug-config.toml"

# Reproduce issue with minimal steps
rfh registry add debug https://registry.example.com
rfh auth login
# ... reproduce issue
```

## Reporting Issues

When creating GitHub issues:

1. **Search existing issues** first
2. **Include system information** (OS, RFH version, etc.)
3. **Provide minimal reproduction steps**
4. **Include full error messages** (use verbose mode)
5. **Describe expected vs actual behavior**
6. **Add relevant configuration** (sanitize tokens!)

### Issue Template

```markdown
**RFH Version**: `rfh --version output`
**OS**: `uname -a` or Windows version
**Command**: `rfh command that failed`

**Expected Behavior**:
What you expected to happen

**Actual Behavior**:
What actually happened

**Error Output**:
```
Full error message with --verbose flag
```

**Steps to Reproduce**:
1. rfh init
2. rfh pack --file=test.mdc
3. Error occurs

**Additional Context**:
Any other relevant information
```

## See Also

- [CLI Commands](../cli/commands.md) - Command reference
- [Configuration Guide](../cli/configuration.md) - Configuration options
- [Installation Guide](installation.md) - Installation instructions
- [GitHub Issues](https://github.com/your-org/rfh/issues) - Report problems