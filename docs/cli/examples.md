# CLI Usage Examples

This document provides common usage patterns and real-world examples of RFH commands.

## Quick Start Workflow

Here's a typical workflow for using RFH:

```bash
# 1. Initialize a new project
mkdir my-rules && cd my-rules
rfh init

# 2. Add dependencies
rfh add security-rules@1.2.0

# 3. Create your rule file
echo "# My Custom Rule" > my-rule.mdc
echo "Never hardcode API keys in source code." >> my-rule.mdc

# 4. Package your rule
rfh pack --file=my-rule.mdc --package=my-custom-rules

# 5. Check staged packages
rfh status

# 6. Publish to registry
rfh publish
```

## Project Setup Examples

### Starting from Scratch
```bash
# Create new project directory
mkdir security-project
cd security-project

# Initialize RFH project
rfh init

# Verify setup
ls -la
# Should show: rulestack.json, CLAUDE.md, .rulestack/
```

### Adding Dependencies
```bash
# Add latest version of a package
rfh add security-rules

# Add specific version
rfh add logging-rules@2.1.0

# Add multiple packages
rfh add security-rules@1.2.0
rfh add performance-rules@1.0.1
rfh add best-practices@2.0.0
```

## Package Creation Examples

### Basic Package Creation
```bash
# Create a rule file
cat << EOF > auth-rules.mdc
# Authentication Rules

## Rule 1: Password Complexity
All passwords must be at least 12 characters long with mixed case, numbers, and symbols.

## Rule 2: Two-Factor Authentication  
All admin accounts must have 2FA enabled.
EOF

# Package the rule
rfh pack --file=auth-rules.mdc --package=auth-security-rules

# Check what was created
rfh status
```

### Updating Existing Packages
```bash
# Create additional rule
echo "# New Authentication Rule" > new-auth.mdc
echo "API keys must be rotated every 90 days." >> new-auth.mdc

# Add to existing package (version auto-increments)
rfh pack --file=new-auth.mdc --package=auth-security-rules

# Or specify version explicitly
rfh pack --file=new-auth.mdc --package=auth-security-rules --version=1.2.0
```

### Custom Output Locations
```bash
# Package with custom output path
rfh pack --file=rules.mdc --package=my-rules --output=releases/my-rules-v1.tgz

# Package to different directory
mkdir exports
rfh pack --file=rules.mdc --package=my-rules --output=exports/
```

## Registry Management Examples

### Setting Up Multiple Registries
```bash
# Add company registry
rfh registry add company https://registry.company.com

# Add public registry
rfh registry add public https://public.rulestack.com

# List all registries
rfh registry list
# Output:
# * company - https://registry.company.com (active)
#   public - https://public.rulestack.com

# Switch between registries
rfh registry use public
rfh registry use company
```

### Working with Different Environments
```bash
# Development environment
rfh registry add dev https://dev-registry.company.com
rfh registry use dev
rfh auth login --username dev-user

# Staging environment  
rfh registry add staging https://staging-registry.company.com
rfh registry use staging
rfh auth login --username staging-user

# Production environment
rfh registry add prod https://registry.company.com
rfh registry use prod
rfh auth login --username prod-user
```

## Authentication Examples

### Interactive Authentication
```bash
# Login with prompts
rfh auth login
# Enter username: myuser
# Enter password: [hidden]
# ✅ Successfully logged in to https://registry.company.com

# Check auth status
rfh auth status
# Logged in as: myuser
# Registry: https://registry.company.com
```

### Non-Interactive Authentication
```bash
# Login with credentials (CI/CD)
rfh auth login --username=$RFH_USER --password=$RFH_PASS

# Login with token
rfh auth login --token=$RFH_TOKEN

# Register new account
rfh auth register --username=newuser --email=user@company.com --password=secure123
```

## Package Discovery Examples

### Searching Packages
```bash
# List all available packages
rfh search

# Search for security-related packages
rfh search security

# Search for specific package
rfh search logging-rules

# Verbose search (shows more details)
rfh search security --verbose
```

### Installing Specific Versions
```bash
# Find available versions
rfh search security-rules --verbose
# Shows: security-rules@1.0.0, security-rules@1.1.0, security-rules@1.2.0

# Install specific version
rfh add security-rules@1.1.0

# Install latest (default)
rfh add security-rules
```

## Advanced Workflows

### Multi-Package Development
```bash
# Working on multiple related packages
rfh init

# Create base rules
rfh pack --file=base.mdc --package=base-rules

# Create extension rules that depend on base
rfh pack --file=web.mdc --package=web-rules 
rfh pack --file=api.mdc --package=api-rules

# Publish all at once
rfh publish
```

### Version Management
```bash
# Check current package versions
rfh status

# Create patch release (auto-increment)
echo "# Minor fix" > hotfix.mdc  
rfh pack --file=hotfix.mdc --package=my-rules
# Output: ✅ Updated existing package: my-rules v1.0.0 -> v1.0.1

# Create minor release
rfh pack --file=feature.mdc --package=my-rules --version=1.1.0

# Create major release  
rfh pack --file=breaking.mdc --package=my-rules --version=2.0.0
```

### CI/CD Integration
```bash
#!/bin/bash
# ci-publish.sh - Example CI script

set -e

# Authenticate
rfh auth login --username=$RFH_CI_USER --password=$RFH_CI_PASS

# Package rules
rfh pack --file=rules.mdc --package=ci-rules --version=$BUILD_VERSION

# Publish
rfh publish

# Logout
rfh auth logout
```

### Bulk Operations
```bash
# Package multiple files
for file in rules/*.mdc; do
  name=$(basename "$file" .mdc)
  rfh pack --file="$file" --package="$name-rules"
done

# Publish everything
rfh publish
```

## Troubleshooting Examples

### Common Issues and Solutions

#### Authentication Problems
```bash
# Check auth status
rfh auth status

# Re-authenticate if needed
rfh auth logout
rfh auth login

# Check registry configuration
rfh registry list
```

#### Package Issues
```bash
# Clear staged packages
rm -rf .rulestack/staged/*

# Rebuild package
rfh pack --file=rules.mdc --package=my-rules

# Check file format
file rules.mdc  # Should be: rules.mdc: UTF-8 Unicode text
```

#### Registry Connection Issues
```bash
# Test registry connectivity
curl -I https://registry.example.com/health

# Try different registry
rfh registry use backup-registry
rfh publish
```

## Environment Variables

RFH supports these environment variables:

```bash
# Registry configuration
export RFH_REGISTRY_URL="https://my-registry.com"
export RFH_AUTH_TOKEN="your-token-here"

# Config file location
export RFH_CONFIG_PATH="/custom/path/config.toml"

# Debug mode
export RFH_DEBUG=1
rfh pack --file=rules.mdc --package=debug-test
```

## See Also

- [Command Reference](commands.md) - Complete command documentation
- [Configuration Guide](configuration.md) - Configuration file reference
- [Troubleshooting](../deployment/troubleshooting.md) - Common issues and solutions