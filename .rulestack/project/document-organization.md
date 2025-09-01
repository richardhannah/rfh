# Document Organization Rule

## Rule: Organize Project Documentation in docs/ Folder

### Overview
All project documentation must be organized in a structured manner to maintain readability and discoverability as the project grows.

### Document Organization Structure

```
project-root/
├── README.md                    # Main project index - links to all other docs
├── docs/                        # All detailed documentation
│   ├── cli/                     # CLI reference and usage
│   │   ├── commands.md          # Command reference
│   │   ├── examples.md          # Usage examples
│   │   └── configuration.md     # Config file reference
│   ├── development/             # Development and contributor guides
│   │   ├── setup.md             # Development environment setup
│   │   ├── testing.md           # Testing guidelines and procedures
│   │   ├── architecture.md      # System architecture overview
│   │   └── contributing.md      # Contribution guidelines
│   ├── api/                     # API documentation (if applicable)
│   │   ├── reference.md         # API reference
│   │   └── authentication.md    # Auth documentation
│   └── deployment/              # Deployment and operations
│       ├── installation.md      # Installation instructions
│       └── troubleshooting.md   # Common issues and solutions
```

### Requirements

#### 1. Main README.md
- **Acts as project index** - provides overview and links to detailed docs
- **Keep concise** - focus on quick start and navigation
- **Link to docs/** - all detailed content goes in docs folder
- **Include project status** - badges, version info, etc.

#### 2. Documentation Categories
Organize docs by these subject areas:

- **CLI Reference** (`docs/cli/`) - Command documentation, examples, configuration
- **Development** (`docs/development/`) - Setup, testing, architecture, contributing
- **API Documentation** (`docs/api/`) - API reference and authentication (if applicable)  
- **Deployment** (`docs/deployment/`) - Installation, operations, troubleshooting

#### 3. Document Naming
- Use **kebab-case** for filenames: `setup.md`, `api-reference.md`
- Use **descriptive names** that clearly indicate content
- Keep filenames **concise but clear**

#### 4. Document Structure
Each document should have:
- **Clear title** (H1 header)
- **Table of contents** for longer documents
- **Consistent formatting** using Markdown
- **Cross-references** to related documents

### Implementation Guidelines

#### When Adding New Documentation:
1. **Determine the category** - which subject area does it belong to?
2. **Place in appropriate subfolder** within docs/
3. **Update README.md** to link to the new document
4. **Cross-reference** from related documents

#### When Reorganizing Docs:
1. **Move files to docs/** maintaining category structure
2. **Update all internal links** to reflect new paths
3. **Update README.md** to reflect new organization
4. **Test all links** to ensure they work

### Examples

#### Good README.md Structure:
```markdown
# Project Name

Brief project description and key features.

## Quick Start
[Basic usage example]

## Documentation
- [CLI Reference](docs/cli/commands.md)
- [Development Setup](docs/development/setup.md)
- [Testing Guide](docs/development/testing.md)
- [API Reference](docs/api/reference.md)

## Installation
[Link to detailed installation guide]
```

#### Good Document Cross-Reference:
```markdown
# Testing Guide

For development setup, see [Development Setup](setup.md).
For CLI usage in tests, see [CLI Commands](../cli/commands.md).
```

### Benefits

1. **Improved Discoverability** - logical organization makes docs easy to find
2. **Better Maintainability** - clear structure reduces documentation debt
3. **Consistent Experience** - users know where to look for information
4. **Scalable Growth** - structure supports growing documentation needs
5. **Professional Appearance** - well-organized docs improve project credibility

### Enforcement

During code review, verify that:
- New documentation is placed in appropriate docs/ subfolder
- README.md is updated to link to new docs
- Internal cross-references use correct relative paths
- Document structure follows established patterns

---

**Remember**: Good documentation organization is as important as good code organization!