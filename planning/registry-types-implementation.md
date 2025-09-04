# Registry Types Implementation Plan

## Overview

This plan introduces the concept of registry types to RuleStack (rfh), expanding beyond the current `remote-http` type to support `git` registries. This will enable users to publish packages to git repositories and consume packages from git repositories.

## Current State Analysis

### Existing Implementation
- **Registry Configuration**: `internal/config/cli.go` - `Registry` struct contains URL, Username, JWTToken
- **Registry Commands**: `internal/cli/registry.go` - Add, list, use, remove registries
- **HTTP Client**: `internal/client/client.go` - Handles HTTP-based registry operations (search, publish, download)
- **Current Registry Type**: Implicit `remote-http` - all registries assume HTTP-based API

### Key Files to Modify
- `internal/config/cli.go` - Add registry type field
- `internal/cli/registry.go` - Update commands to handle registry types  
- `internal/client/` - Create registry client abstraction
- New: `internal/client/git.go` - Git registry client implementation
- New: `internal/client/interface.go` - Registry client interface

## Architecture Design

### Registry Type System

```go
type RegistryType string

const (
    RegistryTypeHTTP RegistryType = "remote-http"
    RegistryTypeGit  RegistryType = "git"
)

type Registry struct {
    URL      string       `toml:"url"`
    Type     RegistryType `toml:"type"`
    Username string       `toml:"username,omitempty"`
    JWTToken string       `toml:"jwt_token,omitempty"`  // For HTTP registries
    GitToken string       `toml:"git_token,omitempty"`  // For Git registries (GitHub/GitLab)
}
```

### Client Interface Abstraction

```go
// RegistryClient defines the interface all registry types must implement
type RegistryClient interface {
    SearchPackages(query, tag, target string, limit int) ([]map[string]interface{}, error)
    GetPackage(name string) (map[string]interface{}, error)
    GetPackageVersion(name, version string) (map[string]interface{}, error)
    PublishPackage(manifestPath, archivePath string) (map[string]interface{}, error)
    DownloadBlob(sha256, destPath string) error
    Health() error
}

// NewRegistryClient factory function
func NewRegistryClient(registry config.Registry) (RegistryClient, error) {
    switch registry.Type {
    case config.RegistryTypeHTTP:
        return NewHTTPClient(registry.URL, registry.JWTToken), nil
    case config.RegistryTypeGit:
        return NewGitClient(registry.URL, registry.GitToken)
    default:
        return nil, fmt.Errorf("unsupported registry type: %s", registry.Type)
    }
}
```

## Git Registry Implementation

### Dependencies
- `github.com/go-git/go-git/v5` - Core Git operations
- `github.com/go-git/go-git/v5/plumbing/transport/http` - HTTP authentication
- Standard library `net/http` - GitHub API for PR creation

### Git Registry Structure
Git registries will follow this directory structure:
```
repo-root/
├── packages/
│   ├── package-name/
│   │   ├── versions/
│   │   │   ├── v1.0.0/
│   │   │   │   ├── manifest.json
│   │   │   │   └── archive.tar.gz
│   │   │   └── v1.1.0/
│   │   │       ├── manifest.json
│   │   │       └── archive.tar.gz
│   │   └── latest.json  # Points to latest version
│   └── another-package/
└── index.json  # Registry metadata
```

### Authentication Strategy

#### Public Repositories
- No authentication required for cloning/reading
- GitHub token required for creating PRs and pushing to forks

#### Private Repositories  
- GitHub token required for all operations
- Use `go-git` HTTP BasicAuth with token

```go
auth := &http.BasicAuth{
    Username: "token", // GitHub uses "token" as username
    Password: token,   // The actual GitHub token
}
```

### Git Registry Operations

#### Search Packages
1. Clone/pull latest version of registry repository
2. Parse `index.json` for package metadata
3. Filter and return results matching query parameters

#### Get Package/Version
1. Clone/pull registry repository  
2. Navigate to `packages/{name}/` directory
3. Read version metadata from `versions/{version}/manifest.json`

#### Publish Package (PR Workflow)
1. Fork the target registry repository (if not already forked)
2. Clone the forked repository
3. Create new branch: `publish/{package-name}/{version}`
4. Add package files to appropriate directory structure
5. Commit changes with descriptive message
6. Push branch to forked repository
7. Create pull request using GitHub API
8. Return PR URL and metadata

#### Download Blob
1. Clone/pull registry repository
2. Locate archive file in package version directory
3. Copy file to destination path

### Error Handling
- **Network Issues**: Retry with exponential backoff
- **Authentication Failures**: Clear error messages about token requirements
- **Repository Access**: Distinguish between public/private access issues
- **PR Creation Failures**: Provide GitHub API error details

## Migration Strategy

### Backward Compatibility
Following the "no backward compatibility" rule, we will:
1. Add `Type` field to `Registry` struct with `remote-http` as default
2. Automatically migrate existing registries to `remote-http` type
3. Update commands to require explicit type specification for new registries

### Configuration Migration
```go
// During config load, migrate old format:
if registry.Type == "" {
    registry.Type = RegistryTypeHTTP
}
```

## Updated CLI Commands

### Registry Add Command
```bash
# HTTP registry (existing behavior)
rfh registry add public https://registry.rulestack.dev --type remote-http

# Git registry 
rfh registry add github-public https://github.com/company/rulestack-registry --type git
rfh registry add github-private https://github.com/company/private-registry --type git --token ghp_xxx
```

### Registry List Command
Enhanced to show registry types:
```
📋 Configured registries:

* public (remote-http)
    URL: https://registry.rulestack.dev
    JWT Token: [configured]

  github-repo (git)
    URL: https://github.com/company/rulestack-registry  
    Git Token: [configured]
```

## Implementation Phases

### Phase 1: Core Architecture
- Add `Type` field to `Registry` struct
- Create `RegistryClient` interface
- Refactor existing HTTP client to implement interface
- Update registry commands to handle types

### Phase 2: Git Client Implementation  
- Implement basic Git client with clone/pull operations
- Add search and get operations
- Implement authentication handling

### Phase 3: Git Publishing (PR Workflow)
- Implement repository forking logic
- Add branch creation and commit functionality  
- Integrate GitHub API for PR creation
- Add comprehensive error handling

### Phase 4: Testing & Refinement
- Add comprehensive test coverage
- Test with public and private repositories
- Performance optimization for large repositories
- Documentation updates

## Key Implementation Considerations

### Performance
- **Repository Caching**: Cache cloned repositories locally to avoid repeated clones
- **Partial Clones**: Use shallow clones when possible to reduce bandwidth
- **Index Optimization**: Keep package index in memory for faster searches

### Security
- **Token Storage**: Securely store GitHub tokens in config
- **Repository Verification**: Validate repository URLs and structure
- **Sandbox Operations**: Perform git operations in temporary directories

### User Experience
- **Clear Error Messages**: Distinguish between authentication, network, and repository issues
- **Progress Indicators**: Show progress for long-running git operations
- **PR Status**: Provide clear feedback on PR creation success/failure

## Testing Strategy

### Unit Tests
- Registry client interface compliance
- Git operations (clone, branch, commit, push)
- GitHub API integration
- Configuration migration

### Integration Tests  
- End-to-end publish workflow with test repositories
- Authentication with various token types
- Error handling for network and API failures

### Manual Testing
- Test with public GitHub repositories
- Test with private repositories
- Verify PR creation and workflow
- Performance testing with large repositories

## Benefits

1. **Decentralized Package Distribution**: Enables package hosting on any git platform
2. **Version Control Integration**: Leverages git's native version control for package management
3. **Open Source Friendly**: Allows community contributions via standard PR workflow
4. **Cost Effective**: No additional infrastructure required for package hosting
5. **Audit Trail**: Full history of package changes through git commit history

## Risks and Mitigations

### Risk: Large Repository Performance
**Mitigation**: Implement shallow clones, repository caching, and selective fetching

### Risk: GitHub API Rate Limits  
**Mitigation**: Implement retry logic, rate limit handling, and user guidance on token requirements

### Risk: Complex Authentication Setup
**Mitigation**: Provide clear documentation and error messages for token configuration

### Risk: Network Reliability
**Mitigation**: Robust retry logic and offline/cached operation modes

## Files to Create/Modify

### New Files
- `internal/client/interface.go` - Registry client interface
- `internal/client/git.go` - Git registry implementation
- `internal/client/factory.go` - Client factory function

### Modified Files
- `internal/config/cli.go` - Add registry type and git token fields
- `internal/cli/registry.go` - Update commands for registry types
- `internal/client/client.go` - Rename to `http.go`, implement interface
- All CLI commands using registry - Update to use client factory

This implementation will provide a robust foundation for git-based package registries while maintaining clean separation between registry types and enabling future extensibility.