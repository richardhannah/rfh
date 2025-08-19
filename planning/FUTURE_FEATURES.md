# RFH Future Features

This document outlines CLI commands and features that were planned but not yet implemented. These may be added in future versions based on user needs and feedback.

## Unimplemented CLI Commands

### `rfh list` / `rfh ls`

**Purpose**: List all installed rulesets in the current workspace

**Planned Functionality**:
- Scan the `.rulestack/` directory to find all installed packages
- Display package information in a clean table format:
  - Package name (with scope if applicable)
  - Version
  - Description
  - Installation date
  - Status (installed/applied)
- Support filtering and sorting options
- Show dependency information
- Indicate which packages are currently "applied" to editors

**Example Usage**:
```bash
rfh list
rfh ls --format=json
rfh list --applied-only
rfh list --sort=name
```

**Example Output**:
```
NAME                    VERSION  STATUS    DESCRIPTION
@acme/security-rules    1.2.0    applied   Security best practices
monty-python-quotes     1.0.0    installed Humorous AI responses
@company/style-guide    2.1.0    applied   Code style guidelines
```

**Implementation Notes**:
- Read from `rulestack.lock.json` for authoritative package list
- Parse manifest files for descriptions and metadata
- Check editor-specific configuration to determine "applied" status
- Support multiple output formats (table, JSON, YAML)

---

### `rfh apply`

**Purpose**: Apply an installed ruleset to specific editor workspaces

**Planned Functionality**:
- Take an installed package and make it active in editor environments
- Support multiple editors (Cursor, Claude Code, Windsurf, VS Code with Copilot)
- Handle editor-specific rule formats and locations
- Manage conflicts between different rulesets
- Allow selective application of rules from a package
- Maintain state of what's applied where

**Example Usage**:
```bash
rfh apply security-rules
rfh apply @acme/style-guide --editor=cursor
rfh apply monty-python-quotes --workspace=./project
rfh apply rules-package --rule="specific-rule.md"
```

**Planned Features**:
- **Multi-Editor Support**: Different editors store AI rules in different formats and locations
- **Workspace Isolation**: Apply different rules to different project workspaces
- **Rule Selection**: Choose specific rules from a package instead of applying all
- **Conflict Resolution**: Handle overlapping or conflicting rules between packages
- **Editor Detection**: Automatically detect which editors are installed and configured

**Editor Integration Points**:
- **Cursor**: Copy rules to `.cursor/` directory
- **Claude Code**: Integration with Claude Code's rule system
- **Windsurf**: Copy to Windsurf configuration directory
- **VS Code + Copilot**: Convert to appropriate format for Copilot

**Implementation Notes**:
- Need editor-specific adapters for different rule formats
- Maintain application state in workspace-specific config files
- Support for rule templates and variable substitution
- Rollback capability to unapply rules
- Integration with editor configuration systems

---

## Additional Future Features

### Enhanced Package Management

**Dependency Resolution**:
- Support for package dependencies in `rulestack.json`
- Automatic installation of required dependencies
- Version constraint resolution
- Dependency tree visualization

**Package Versions**:
- Support for version ranges (e.g., `^1.0.0`, `~2.1.0`)
- Automatic updates within semantic version constraints
- Version conflict detection and resolution

**Workspace Management**:
- Project-specific rule configurations
- Rule inheritance from parent directories
- Environment-specific rule sets (dev/prod/test)

### Advanced CLI Features

**Interactive Mode**:
- `rfh interactive` - GUI-like experience for package management
- Rule browser and preview functionality
- Visual dependency management

**Rule Development Tools**:
- `rfh validate` - Validate rule syntax and effectiveness
- `rfh test` - Test rules against sample code
- `rfh generate` - Generate rule templates

**Registry Management**:
- Private registry support with authentication
- Registry mirroring and caching
- Offline mode with local package cache

### Integration Features

**IDE Plugins**:
- VS Code extension for RFH management
- Cursor integration plugin
- Real-time rule application and preview

**CI/CD Integration**:
- GitHub Actions for automated rule testing
- Integration with code review systems
- Automated rule compliance checking

**Collaboration Features**:
- Team rule sharing and synchronization
- Rule usage analytics and recommendations
- Community rule marketplace

---

## Implementation Priority

When considering future implementation, the suggested order would be:

1. **`rfh list`** - Fundamental for usability, helps users see what they have installed
2. **`rfh apply`** - Core functionality for actually using the rules
3. **Dependency Resolution** - Important for complex rule ecosystems
4. **IDE Integrations** - Enhanced user experience
5. **Advanced Features** - Polish and convenience features

---

## Notes

- These features were designed but not implemented in the initial version
- Implementation should be driven by user feedback and actual usage patterns
- Each feature should maintain the simplicity and clarity of the current CLI design
- Backward compatibility should be maintained when adding new features

---

*This document should be updated as new features are planned or existing plans change.*