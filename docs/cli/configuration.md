# Configuration Guide

This document covers RFH configuration options and file formats.

## Configuration File Location

RFH uses a TOML configuration file located at:

- **Linux/Mac**: `~/.rfh/config.toml`
- **Windows**: `%USERPROFILE%\.rfh\config.toml`

The configuration file is created automatically on first use.

## Configuration File Format

### Basic Structure

```toml
[registry]
active = "default"

[[registries]]
name = "default"
url = "https://registry.example.com"

[[registries]]
name = "staging"
url = "https://staging.example.com"

[auth]
token = "your-auth-token-here"
```

### Registry Configuration

#### Active Registry
```toml
[registry]
active = "production"  # Name of the active registry
```

#### Registry List
```toml
[[registries]]
name = "production"
url = "https://registry.company.com"

[[registries]]
name = "staging"
url = "https://staging.company.com"

[[registries]]
name = "local"
url = "http://localhost:8080"
```

**Registry Fields:**
- `name` (string) - Unique identifier for the registry
- `url` (string) - Base URL of the registry API

### Authentication Configuration

```toml
[auth]
token = "eyJhbGciOiJIUzI1NiIs..."  # Bearer token for active registry
```

The auth token is automatically managed when you run `rfh auth login`.

## Environment Variables

RFH supports these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `RFH_CONFIG_PATH` | Path to config file | `~/.rfh/config.toml` |
| `RFH_REGISTRY_URL` | Override active registry URL | - |
| `RFH_AUTH_TOKEN` | Override auth token | - |
| `RFH_DEBUG` | Enable debug logging | `false` |

### Examples

```bash
# Use custom config location
export RFH_CONFIG_PATH="/opt/rfh/config.toml"

# Override registry for CI/CD
export RFH_REGISTRY_URL="https://ci-registry.company.com"
export RFH_AUTH_TOKEN="$CI_AUTH_TOKEN"

# Enable debug mode
export RFH_DEBUG=1
rfh publish --verbose
```

## Command-Line Overrides

Global flags can override configuration settings:

```bash
# Override registry
rfh search --registry=https://other-registry.com

# Override auth token  
rfh publish --token=custom-token

# Use different config file
rfh --config=/tmp/test-config.toml init
```

## Project Manifest (rulestack.json)

Each RFH project contains a `rulestack.json` file with project-specific configuration.

### Project Manifest Format

```json
{
  "version": "1.0.0",
  "dependencies": {
    "security-rules": "1.2.0",
    "logging-rules": "2.1.0",
    "best-practices": "1.0.1"
  }
}
```

**Project Manifest Fields:**
- `version` (string) - Project version
- `dependencies` (object) - Map of package names to versions

### Dependency Management

Dependencies are automatically managed when you run:

```bash
# Add dependency
rfh add security-rules@1.2.0
# Adds to dependencies in rulestack.json

# Install all dependencies
rfh install
# Installs packages listed in dependencies
```

## Advanced Configuration

### Multiple Environments

You can manage different environments using separate config files:

```bash
# Development
RFH_CONFIG_PATH=~/.rfh/dev.toml rfh auth login

# Staging  
RFH_CONFIG_PATH=~/.rfh/staging.toml rfh auth login

# Production
RFH_CONFIG_PATH=~/.rfh/prod.toml rfh auth login
```

### CI/CD Configuration

For automated environments:

```bash
#!/bin/bash
# ci-setup.sh

export RFH_CONFIG_PATH="/tmp/rfh-config.toml"
export RFH_REGISTRY_URL="$CI_REGISTRY_URL"
export RFH_AUTH_TOKEN="$CI_AUTH_TOKEN"

# Commands will use environment variables
rfh publish
```

Or create a minimal config file:

```toml
# /tmp/rfh-config.toml
[registry]
active = "ci"

[[registries]]
name = "ci"
url = "${CI_REGISTRY_URL}"

[auth]
token = "${CI_AUTH_TOKEN}"
```

### Private Registry Setup

For private registries with custom certificates:

```toml
[registry]
active = "private"

[[registries]]
name = "private"
url = "https://private.company.com"
# Note: SSL verification is enabled by default
# Use --insecure flag only for testing

[auth]
token = "private-registry-token"
```

## Security Considerations

### Token Storage

- Auth tokens are stored in the config file
- Config file has restricted permissions (600 on Unix)
- Tokens are not encrypted at rest
- Use environment variables in CI/CD to avoid storing tokens in files

### Registry URLs

- Always use HTTPS in production
- Verify certificate validity
- Use `--insecure` flag only for local development

### Config File Permissions

```bash
# Ensure proper permissions
chmod 600 ~/.rfh/config.toml

# Check permissions
ls -la ~/.rfh/config.toml
# Should show: -rw------- (owner read/write only)
```

## Configuration Examples

### Development Setup
```toml
[registry]
active = "local"

[[registries]]
name = "local"
url = "http://localhost:8080"

[[registries]]
name = "dev"
url = "https://dev-registry.company.com"

[auth]
token = "dev-token-12345"
```

### Production Setup
```toml
[registry]
active = "prod"

[[registries]]
name = "prod"
url = "https://registry.company.com"

[[registries]]
name = "staging"
url = "https://staging.company.com"

[auth]
token = "prod-token-secure-hash"
```

### Multi-tenant Setup
```toml
[registry]
active = "tenant-a"

[[registries]]
name = "tenant-a"
url = "https://tenant-a.registry.com"

[[registries]]
name = "tenant-b"
url = "https://tenant-b.registry.com"

[[registries]]
name = "shared"
url = "https://shared.registry.com"

[auth]
token = "tenant-specific-token"
```

## Troubleshooting Configuration

### Common Issues

#### Config File Not Found
```bash
# Check if config directory exists
ls -la ~/.rfh/

# Create directory if missing
mkdir -p ~/.rfh

# RFH will create config.toml on first use
rfh registry add default https://registry.example.com
```

#### Permission Denied
```bash
# Fix permissions
chmod 600 ~/.rfh/config.toml
chmod 700 ~/.rfh
```

#### Invalid TOML Format
```bash
# Validate TOML syntax
# Use online validator or:
rfh registry list
# Will show TOML parsing errors
```

#### Registry Connection Issues
```bash
# Test registry connectivity
curl -I https://registry.example.com/v1/health

# Check configuration
rfh registry list

# Verify auth token
rfh auth status
```

### Debug Configuration

Enable verbose output to debug configuration issues:

```bash
# Show configuration resolution
RFH_DEBUG=1 rfh --verbose registry list

# Show effective configuration
rfh --verbose auth status
```

## Migrating Configuration

### From Older Versions

If upgrading from an older version of RFH:

1. **Backup existing config:**
   ```bash
   cp ~/.rfh/config.toml ~/.rfh/config.toml.backup
   ```

2. **Check for deprecated settings:**
   ```bash
   rfh registry list
   # Will warn about deprecated configuration
   ```

3. **Update format if needed:**
   ```bash
   # RFH will automatically migrate on next run
   rfh auth status
   ```

### Reset Configuration

To start fresh:

```bash
# Remove config directory
rm -rf ~/.rfh

# Reconfigure from scratch
rfh registry add default https://registry.example.com
rfh auth login
```

## See Also

- [CLI Commands](commands.md) - Command reference
- [CLI Examples](examples.md) - Usage examples  
- [Installation Guide](../deployment/installation.md) - Installation instructions
- [Troubleshooting](../deployment/troubleshooting.md) - Common issues