# CLI Command Reference

This document provides a comprehensive reference for all RFH commands.

## Global Options

All commands support these global options:

- `--config string` - Config file (default: `~/.rfh/config.toml`)
- `--registry string` - Registry URL override
- `--token string` - Auth token override
- `-v, --verbose` - Verbose output

## Commands Overview

| Command | Purpose |
|---------|---------|
| `rfh init` | Initialize a new RuleStack project |
| `rfh add <package>` | Add a package dependency |
| `rfh pack` | Package rules into distributable archive |
| `rfh publish` | Publish package to registry |
| `rfh search [query]` | Search for packages |
| `rfh status` | Show staged packages |
| `rfh registry` | Manage registries |
| `rfh auth` | Authentication commands |

---

## Project Management

### `rfh init`

Initialize a new RuleStack project in the current directory.

**Usage:**
```bash
rfh init [flags]
```

**Flags:**
- `-f, --force` - Force overwrite existing files

**Creates:**
- `rulestack.json` - Project manifest file
- `.rulestack/` - Dependency directory with core rules
- `CLAUDE.md` - Claude Code integration file

**Example:**
```bash
rfh init
# ✅ Initialized RuleStack project in: my-project
```

### `rfh status`

Show staged packages ready for publishing.

**Usage:**
```bash
rfh status
```

**Example:**
```bash
rfh status
# Staged packages:
# - security-rules-1.2.0.tgz
# - logging-rules-1.0.1.tgz
```

---

## Package Management

### `rfh add <package>`

Add a package dependency to your project.

**Usage:**
```bash
rfh add <package>[@version] [flags]
```

**Examples:**
```bash
# Add latest version
rfh add security-rules

# Add specific version
rfh add security-rules@1.2.0

# Add with verbose output
rfh add security-rules --verbose
```

### `rfh pack`

Package rule files into a distributable archive.

**Usage:**
```bash
rfh pack [flags]
```

**Flags:**
- `-f, --file string` - .mdc file to pack (required)
- `-o, --output string` - Output archive path
- `-p, --package string` - Package name (enables non-interactive mode)
- `--version string` - Package version (auto-increments for existing packages, defaults to 1.0.0 for new packages)

**Examples:**
```bash
# Interactive mode - prompts for package details
rfh pack --file=rules.mdc

# Non-interactive - specify package name
rfh pack --file=rules.mdc --package=my-rules

# Specify version explicitly
rfh pack --file=rules.mdc --package=my-rules --version=2.1.0

# Custom output path
rfh pack --file=rules.mdc --package=my-rules --output=custom.tgz
```

**Enhanced Pack Features:**
- **Existing Package Detection** - Automatically detects existing packages
- **Version Auto-increment** - Bumps patch version for existing packages
- **File Aggregation** - Combines existing package files with new files
- **Version Validation** - Prevents version decreases

### `rfh publish`

Publish staged packages to the registry.

**Usage:**
```bash
rfh publish [flags]
```

**Examples:**
```bash
# Publish all staged packages
rfh publish

# Publish with registry override
rfh publish --registry=https://my-registry.com
```

### `rfh search`

Search for packages in the registry.

**Usage:**
```bash
rfh search [query] [flags]
```

**Examples:**
```bash
# List all packages
rfh search

# Search for specific packages
rfh search security

# Search with verbose output
rfh search security --verbose
```

---

## Registry Management

### `rfh registry`

Manage package registries.

**Subcommands:**
- `add <name> <url>` - Add a new registry
- `list` - List all configured registries
- `use <name>` - Set active registry
- `remove <name>` - Remove a registry

**Examples:**
```bash
# Add registry
rfh registry add myregistry https://registry.example.com

# List registries
rfh registry list

# Switch active registry
rfh registry use myregistry

# Remove registry
rfh registry remove myregistry
```

---

## Authentication

### `rfh auth`

Authentication commands.

**Subcommands:**
- `login` - Login to active registry
- `logout` - Logout from active registry
- `register` - Register new account
- `status` - Show authentication status

**Examples:**
```bash
# Interactive login
rfh auth login

# Non-interactive login
rfh auth login --username=myuser --password=mypass

# Check authentication status
rfh auth status

# Logout
rfh auth logout

# Register new account
rfh auth register --username=newuser --email=user@example.com
```

---

## File Formats

### .mdc Files
Rule files must use the `.mdc` extension and contain valid markdown content.

### rulestack.json
Project manifest file with dependency information:
```json
{
  "version": "1.0.0",
  "projectRoot": "/path/to/project", 
  "dependencies": {
    "security-rules": "1.2.0",
    "logging-rules": "1.0.1"
  }
}
```

### Configuration
Config file location: `~/.rfh/config.toml`

```toml
[registry]
active = "default"

[[registries]]
name = "default"
url = "https://registry.example.com"

[auth]
token = "your-auth-token"
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication required |
| 3 | Package not found |
| 4 | Version validation failed |
| 5 | File validation failed |

---

## See Also

- [CLI Examples](examples.md) - Common usage patterns
- [Configuration Guide](configuration.md) - Detailed configuration reference
- [Development Setup](../development/setup.md) - Development environment setup