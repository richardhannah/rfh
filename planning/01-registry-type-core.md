# Phase 1: Registry Type Core Architecture

## Overview
Add registry type support to the configuration system, establishing the foundation for multiple registry implementations.

## Scope
- Add `Type` field to Registry struct
- Define registry type constants
- Update configuration loading/saving
- Modify registry CLI commands to handle types
- Ensure backward compatibility with existing registries

## Implementation Steps

### 1. Update Configuration Structure

**File**: `internal/config/cli.go`

Add registry type definitions:
```go
type RegistryType string

const (
    RegistryTypeHTTP RegistryType = "remote-http"
    RegistryTypeGit  RegistryType = "git"
)

type Registry struct {
    URL      string       `toml:"url"`
    Type     RegistryType `toml:"type"`        // New field
    Username string       `toml:"username,omitempty"`
    JWTToken string       `toml:"jwt_token,omitempty"`
    GitToken string       `toml:"git_token,omitempty"` // New field for git auth
}
```

### 2. Configuration Migration Logic

**File**: `internal/config/cli.go`

Update `LoadCLI()` function:
```go
func LoadCLI() (CLIConfig, error) {
    // ... existing loading code ...
    
    // Migrate existing registries to have explicit type
    for name, reg := range config.Registries {
        if reg.Type == "" {
            reg.Type = RegistryTypeHTTP
            config.Registries[name] = reg
        }
    }
    
    return config, nil
}
```

### 3. Update Registry Add Command

**File**: `internal/cli/registry.go`

Modify `registryAddCmd`:
```go
var registryAddCmd = &cobra.Command{
    Use:   "add <name> <url> [--type remote-http|git]",
    Short: "Add a new registry",
    Long: `Add a new registry configuration.
    
Registry Types:
  remote-http - Traditional HTTP-based registry (default)
  git        - Git repository-based registry

Examples:
  rfh registry add public https://registry.rulestack.dev
  rfh registry add github https://github.com/org/registry --type git`,
    Args: cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        url := args[1]
        registryType, _ := cmd.Flags().GetString("type")
        
        if registryType == "" {
            registryType = string(RegistryTypeHTTP)
        }
        
        return runRegistryAdd(name, url, RegistryType(registryType))
    },
}

func init() {
    registryAddCmd.Flags().String("type", "remote-http", "Registry type (remote-http or git)")
}
```

### 4. Update Registry Add Implementation

**File**: `internal/cli/registry.go`

```go
func runRegistryAdd(name, url string, registryType RegistryType) error {
    // Validate registry type
    if registryType != RegistryTypeHTTP && registryType != RegistryTypeGit {
        return fmt.Errorf("invalid registry type: %s", registryType)
    }
    
    // Validate URL based on type
    if registryType == RegistryTypeGit {
        if !strings.HasPrefix(url, "https://github.com/") && 
           !strings.HasPrefix(url, "https://gitlab.com/") &&
           !strings.HasPrefix(url, "git@") {
            fmt.Printf("⚠️  Warning: Git registry URL may not be valid\n")
        }
    }
    
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Add registry with type
    cfg.Registries[name] = config.Registry{
        URL:  url,
        Type: registryType,
    }
    
    // ... rest of existing logic ...
}
```

### 5. Update Registry List Command

**File**: `internal/cli/registry.go`

```go
func runRegistryList() error {
    // ... existing code ...
    
    for name, reg := range cfg.Registries {
        marker := "  "
        if cfg.Current == name {
            marker = "* "
        }
        
        registryType := reg.Type
        if registryType == "" {
            registryType = RegistryTypeHTTP
        }
        
        fmt.Printf("%s%s (%s)\n", marker, name, registryType)
        fmt.Printf("    URL: %s\n", reg.URL)
        
        // Show appropriate token status based on type
        if registryType == RegistryTypeHTTP && reg.JWTToken != "" {
            fmt.Printf("    JWT Token: [configured]\n")
        } else if registryType == RegistryTypeGit && reg.GitToken != "" {
            fmt.Printf("    Git Token: [configured]\n")
        }
        
        fmt.Printf("\n")
    }
    
    // ... rest of existing logic ...
}
```

### 6. Add Type Validation Helper

**File**: `internal/config/cli.go`

```go
// ValidateRegistryType checks if a registry type is valid
func ValidateRegistryType(t RegistryType) error {
    switch t {
    case RegistryTypeHTTP, RegistryTypeGit:
        return nil
    default:
        return fmt.Errorf("unsupported registry type: %s", t)
    }
}

// GetEffectiveType returns the effective type for a registry
func (r Registry) GetEffectiveType() RegistryType {
    if r.Type == "" {
        return RegistryTypeHTTP
    }
    return r.Type
}
```

## Testing Requirements

### Unit Tests
1. Test configuration migration (empty type → remote-http)
2. Test registry type validation
3. Test registry add with different types
4. Test backward compatibility

### Integration Tests
1. Add remote-http registry
2. Add git registry
3. List registries showing types
4. Load existing config without types

### Manual Testing Checklist
- [ ] Existing registries without type work correctly
- [ ] Can add remote-http registry explicitly
- [ ] Can add git registry
- [ ] Registry list shows correct types
- [ ] Invalid registry types are rejected
- [ ] Config file correctly saves registry types

### Cucumber Test Amendments

**File**: `features/registry-management.feature`

Add new scenarios for registry types:
```gherkin
  Scenario: Add HTTP registry with explicit type
    When I run "rfh registry add http-typed https://registry.example.com --type remote-http"
    Then the command should succeed
    And the config file should contain registry "http-typed" with type "remote-http"
    
  Scenario: Add Git registry
    When I run "rfh registry add git-registry https://github.com/org/registry --type git"
    Then the command should succeed
    And the config file should contain registry "git-registry" with type "git"
    
  Scenario: Add registry without type defaults to HTTP
    When I run "rfh registry add default-registry https://registry.example.com"
    Then the command should succeed
    And the config file should contain registry "default-registry" with type "remote-http"
    
  Scenario: Reject invalid registry type
    When I run "rfh registry add invalid https://example.com --type invalid-type"
    Then the command should fail
    And the output should contain "invalid registry type"
    
  Scenario: List shows registry types
    Given a registry "typed-http" with URL "https://http.example.com" and type "remote-http"
    And a registry "typed-git" with URL "https://github.com/org/repo" and type "git"
    When I run "rfh registry list"
    Then the output should contain "typed-http (remote-http)"
    And the output should contain "typed-git (git)"
```

**File**: `features/step_definitions/registry_steps.js`

Add new step definitions:
```javascript
Then('the config file should contain registry {string} with type {string}', 
  async function (registryName, registryType) {
    const config = await this.loadConfig();
    assert(config.registries[registryName], 
      `Registry ${registryName} not found in config`);
    assert.equal(config.registries[registryName].type, registryType,
      `Registry type mismatch: expected ${registryType}, got ${config.registries[registryName].type}`);
});

Given('a registry {string} with URL {string} and type {string}', 
  async function (name, url, type) {
    const config = await this.loadConfig();
    if (!config.registries) config.registries = {};
    config.registries[name] = { url, type };
    await this.saveConfig(config);
});
```

**File**: `features/backward-compatibility.feature` (new file)

```gherkin
Feature: Registry Type Backward Compatibility
  Existing registries without type field should continue working

  Background:
    Given I have a clean test environment

  Scenario: Load config with registries missing type field
    Given a config file with content:
      """
      current = "legacy"
      
      [registries.legacy]
      url = "https://legacy.example.com"
      jwt_token = "token123"
      """
    When I run "rfh registry list"
    Then the command should succeed
    And the output should contain "legacy (remote-http)"
    
  Scenario: Existing registry operations work without type
    Given a config file with content:
      """
      [registries.old]
      url = "https://old.example.com"
      """
    When I run "rfh registry use old"
    Then the command should succeed
```

## Success Criteria
- Existing registries continue to work without modification
- New registries can be added with explicit types
- Registry type is displayed in list command
- Configuration correctly persists registry types
- Type validation prevents invalid registry types

## Dependencies
None - this is the foundation phase

## Risks
- **Risk**: Breaking existing configurations
  **Mitigation**: Auto-migration of missing types to remote-http
  
- **Risk**: Invalid registry types in config
  **Mitigation**: Validation on load and command execution

## Next Phase
Phase 2: Registry Client Interface - Creating the abstraction layer for multiple registry implementations