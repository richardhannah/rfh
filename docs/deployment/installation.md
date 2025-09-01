# Installation Guide

This guide covers how to install RFH on different platforms and environments.

## Prerequisites

- Operating System: Windows 10+, macOS 10.15+, or Linux
- For development: Go 1.21+ and Docker (optional)

## Installation Methods

### Option 1: Download Binary (Recommended)

Download the latest binary for your platform from the releases page.

#### Windows
```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/your-org/rfh/releases/latest/download/rfh-windows.exe" -OutFile "rfh.exe"

# Add to PATH (optional)
Move-Item rfh.exe $env:USERPROFILE\bin\rfh.exe
```

#### macOS
```bash
# Download latest release
curl -L -o rfh "https://github.com/your-org/rfh/releases/latest/download/rfh-macos"
chmod +x rfh

# Add to PATH
sudo mv rfh /usr/local/bin/
```

#### Linux
```bash
# Download latest release
curl -L -o rfh "https://github.com/your-org/rfh/releases/latest/download/rfh-linux"
chmod +x rfh

# Add to PATH
sudo mv rfh /usr/local/bin/
```

### Option 2: Build from Source

#### Prerequisites
- Go 1.21 or later
- Git

#### Build Steps
```bash
# Clone repository
git clone https://github.com/your-org/rfh.git
cd rfh

# Build binary
go build -o dist/rfh ./cmd/cli

# Install to PATH (optional)
sudo cp dist/rfh /usr/local/bin/
```

### Option 3: Package Managers

#### Homebrew (macOS/Linux)
```bash
brew tap your-org/rfh
brew install rfh
```

#### Chocolatey (Windows)
```powershell
choco install rfh
```

#### Snap (Linux)
```bash
snap install rfh
```

## Verification

Verify your installation:

```bash
# Check version
rfh --version
# Output: rfh version 1.0.0

# Check help
rfh --help
# Shows command usage
```

## Configuration

### First-Time Setup

1. **Initialize configuration:**
   ```bash
   # Create config directory
   mkdir -p ~/.rfh
   
   # RFH will create config.toml on first use
   ```

2. **Add a registry:**
   ```bash
   # Add your registry
   rfh registry add myregistry https://registry.example.com
   
   # Set as active
   rfh registry use myregistry
   ```

3. **Authenticate:**
   ```bash
   # Login to registry
   rfh auth login
   ```

### Configuration File

Default location: `~/.rfh/config.toml`

Example configuration:
```toml
[registry]
active = "default"

[[registries]]
name = "default"
url = "https://registry.example.com"

[[registries]]
name = "staging"
url = "https://staging-registry.example.com"

[auth]
token = "your-auth-token"
```

## Docker Installation (Optional)

For development or registry hosting:

```bash
# Clone repository
git clone https://github.com/your-org/rfh.git
cd rfh

# Start with Docker Compose
docker-compose up -d

# Build CLI in container
docker-compose exec api go build -o /dist/rfh ./cmd/cli
```

## Platform-Specific Notes

### Windows

- **PowerShell Execution Policy**: You may need to allow script execution:
  ```powershell
  Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```

- **Windows Defender**: Add exclusion for rfh.exe if needed

- **PATH Configuration**: Add to system PATH through System Properties > Environment Variables

### macOS

- **Gatekeeper**: You may need to allow the unsigned binary:
  ```bash
  sudo spctl --add rfh
  sudo xattr -rd com.apple.quarantine rfh
  ```

- **Homebrew Installation**: Recommended for automatic updates

### Linux

- **Dependencies**: Most distributions include required libraries
- **Permissions**: Ensure execute permissions: `chmod +x rfh`
- **Systemd**: Can create service file for background operations

## Development Installation

### Full Development Environment

```bash
# Clone repository
git clone https://github.com/your-org/rfh.git
cd rfh

# Install dependencies
go mod tidy
cd cucumber-testing && npm install && cd ..

# Start development services
docker-compose up -d

# Build CLI
go build -o dist/rfh ./cmd/cli

# Run tests
./run-tests.sh
```

### IDE Setup

#### VS Code
Recommended extensions:
- Go extension
- Docker extension
- Cucumber (Gherkin) Full Support

#### GoLand/IntelliJ
- Enable Go plugin
- Configure run configurations for tests

## Troubleshooting Installation

### Common Issues

#### Binary Not Found
```bash
# Check PATH
echo $PATH

# Verify binary location
which rfh

# Add to PATH if needed
export PATH=$PATH:/path/to/rfh
```

#### Permission Denied
```bash
# Make executable
chmod +x rfh

# Or run with explicit path
./rfh --version
```

#### SSL Certificate Issues
```bash
# Trust certificates (if needed)
rfh registry add myregistry https://registry.example.com --insecure

# Or configure certificates
export SSL_CERT_FILE=/path/to/cert.pem
```

#### Port Conflicts (Development)
```bash
# Check what's using port 8080
netstat -tulpn | grep 8080

# Use different port
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

### Getting Help

If you encounter issues:

1. **Check the logs:**
   ```bash
   rfh --verbose [command]
   ```

2. **Verify configuration:**
   ```bash
   rfh registry list
   rfh auth status
   ```

3. **Test connectivity:**
   ```bash
   curl -I https://your-registry.com/health
   ```

4. **Report issues:**
   - Check existing issues: https://github.com/your-org/rfh/issues
   - Create new issue with system info and error messages

## Updating RFH

### Binary Installation
Download and replace the binary with the latest version.

### Package Managers
```bash
# Homebrew
brew upgrade rfh

# Chocolatey
choco upgrade rfh

# Snap
snap refresh rfh
```

### Build from Source
```bash
git pull origin main
go build -o dist/rfh ./cmd/cli
```

## Uninstallation

### Remove Binary
```bash
# Remove from PATH
sudo rm /usr/local/bin/rfh

# Or remove from custom location
rm ~/bin/rfh
```

### Remove Configuration
```bash
# Remove config directory
rm -rf ~/.rfh
```

### Package Managers
```bash
# Homebrew
brew uninstall rfh

# Chocolatey
choco uninstall rfh

# Snap
snap remove rfh
```

## Next Steps

After installation:

1. **Read the [CLI Command Reference](../cli/commands.md)**
2. **Try the [CLI Examples](../cli/examples.md)**
3. **Set up your [Development Environment](../development/setup.md)** (if developing)
4. **Join the community** (Discord/Slack links)

## See Also

- [CLI Command Reference](../cli/commands.md)
- [Configuration Guide](../cli/configuration.md)
- [Troubleshooting](troubleshooting.md)
- [Development Setup](../development/setup.md)