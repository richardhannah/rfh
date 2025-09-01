# RuleStack (RFH) - Registry for Humans

[![CI](https://github.com/username/rulestack/actions/workflows/ci.yml/badge.svg)](https://github.com/username/rulestack/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/username/rulestack)](https://goreportcard.com/report/github.com/username/rulestack)
[![codecov](https://codecov.io/gh/username/rulestack/branch/master/graph/badge.svg)](https://codecov.io/gh/username/rulestack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A secure package manager for AI rulesets, making it easy to share and discover AI rules for code editors like Claude Code, Cursor, and Windsurf. Built with enterprise-grade security validation and automatic Claude Code integration.

## 📚 Documentation

### Getting Started
- **[Installation Guide](docs/deployment/installation.md)** - Install RFH on your system
- **[CLI Commands](docs/cli/commands.md)** - Complete command reference  
- **[CLI Examples](docs/cli/examples.md)** - Common usage patterns and workflows

### Development
- **[Development Setup](docs/development/setup.md)** - Set up development environment
- **[Testing Guide](docs/development/testing.md)** - Testing procedures and guidelines

### Reference
- **[Configuration Guide](docs/cli/configuration.md)** - Config file reference
- **[Troubleshooting](docs/deployment/troubleshooting.md)** - Common issues and solutions

## 🚀 Quick Start

### Installation
```bash
# Download latest release for your platform
# See: docs/deployment/installation.md

# Verify installation
rfh --version
```

### Basic Usage
```bash
# Initialize new project
rfh init

# Add a package dependency  
rfh add security-rules@1.2.0

# Create and publish your own package
rfh pack --file=my-rule.mdc --package=my-rules
rfh publish
```

### Developer Setup
```bash
# Clone and start development environment
git clone <repo-url> && cd rfh
docker-compose up -d

# Run comprehensive test suite
./run-tests.ps1  # Windows
# ./run-tests.sh   # Linux/Mac
```

📖 **For detailed instructions, see [Development Setup](docs/development/setup.md)**

## What is RuleStack?

RuleStack is a secure package manager for AI rulesets - configuration files and prompts that guide AI coding assistants. Like npm for JavaScript or pip for Python, RuleStack allows you to:

- **📦 Publish** rulesets to public or private registries
- **🔍 Discover** rulesets created by the community  
- **⚡ Install** rulesets with enterprise-grade security validation
- **🔖 Version** rulesets with semantic versioning
- **👥 Share** best practices across teams and projects
- **🤖 Integrate** automatically with Claude Code via CLAUDE.md

## Key Features

- **🔒 Enterprise Security** - Malware protection, path traversal prevention, content sanitization
- **🤖 AI Editor Integration** - Automatic CLAUDE.md updates, core rules system
- **🚀 Developer Experience** - Git-like workflow with familiar commands
- **📁 Project Structure** - Clear project boundaries like `git init`

## Supported Editors

- **Cursor** - AI-powered code editor
- **Claude Code** - Anthropic's coding assistant  
- **Windsurf** - AI development environment
- **GitHub Copilot** - Microsoft's AI pair programmer

## Architecture

- **CLI & API**: Go + Cobra + PostgreSQL
- **Security**: Comprehensive validation against malicious packages
- **Storage**: Filesystem with integrity verification
- **Integration**: Seamless Claude Code integration

## Contributing

We welcome contributions! See our [Contributing Guidelines](CONTRIBUTING.md) for details.

**Philosophy**: We prioritize working software over perfect code. If it works and improves the project, we want it!

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/your-org/rfh/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/rfh/discussions)

---

**RuleStack** - Making AI rulesets accessible and shareable for everyone. 🚀