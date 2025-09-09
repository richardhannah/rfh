# Git Registry Authentication Workflow

## Current Implementation Analysis

### HTTP Registries (remote-http)
- Use JWT tokens stored in config
- `rfh auth login` - interactive/non-interactive login with username/password
- `rfh auth register` - create account
- Token stored per-registry in config.toml

### Git Registries
- Use `git_token` field in config.toml
- Token passed to GitClient during initialization
- No dedicated auth command for Git registries
- Users must manually add token to config or use GITHUB_TOKEN env var

## Issues with Current Git Registry Auth

1. **No CLI command to set Git tokens** - users must manually edit config.toml
2. **Inconsistent with HTTP auth flow** - HTTP has `auth login`, Git doesn't
3. **No token validation** - can't verify token works until publish attempt
4. **Poor user experience** - unclear how to authenticate

## Proposed Solution: Add Git-specific Auth Commands

### New command: `rfh auth git-token`
- Set GitHub personal access token for current Git registry
- Validate token with GitHub API
- Store in config.toml's `git_token` field
- Support both interactive and non-interactive modes

## Implementation Plan

### 1. Add new subcommand `auth git-token`
- Check if current registry is type "git"
- Prompt for token (or accept via --token flag)
- Validate token using GitHub API (GetAuthenticatedUser)
- Check collaborator access if possible
- Store token in registry config

### 2. Enhance `auth whoami` for Git registries
- Show GitHub username when Git registry is active
- Display token status and permissions

### 3. Update `auth login` behavior
- Detect registry type
- For HTTP: existing username/password flow
- For Git: redirect to `auth git-token` command

### 4. Add token validation during operations
- Validate token before publish attempts
- Provide clear error messages for permission issues

## Files to Modify

- `internal/cli/auth.go` - add git-token command
- `internal/cli/auth.go` - update whoami for Git registries
- `internal/cli/auth.go` - update login to handle Git registries

## Example Usage After Implementation

```bash
# Add Git registry
rfh registry add github-rules https://github.com/myorg/rules-registry --type git

# Set as active
rfh registry use github-rules

# Authenticate (new command)
rfh auth git-token
> Enter GitHub Personal Access Token: ghp_xxxxxxxxxxxx
> ✅ Token validated for user: myusername
> ✅ Collaborator access verified for myorg/rules-registry
> 🔑 Token saved to config

# Or non-interactive
rfh auth git-token --token ghp_xxxxxxxxxxxx

# Check auth status
rfh auth whoami
> 👤 GitHub user: myusername
> 📁 Registry: github-rules (git)
> 🔑 Token: [configured]
> ✅ Collaborator access: verified
```

## Benefits

1. **Consistent UX** - Similar authentication flow for both registry types
2. **Token Validation** - Verify token works before attempting operations
3. **Clear Feedback** - Users know if they're authenticated correctly
4. **Better Error Messages** - Can detect permission issues early
5. **Non-interactive Support** - Automation-friendly with --token flag

## Alternative Considerations

### Option 1: Unified `auth login` command
Instead of separate commands, make `auth login` smart enough to handle both:
- For HTTP: prompt username/password
- For Git: prompt for token

Pros: Single command for all auth
Cons: Different input requirements might confuse users

### Option 2: Environment Variable Only
Keep using GITHUB_TOKEN env var only, no config storage.

Pros: Simple, follows GitHub conventions
Cons: Must set env var every session, can't have different tokens per registry

### Recommendation
Implement the proposed `auth git-token` command as it provides the best balance of usability, flexibility, and consistency with existing patterns.