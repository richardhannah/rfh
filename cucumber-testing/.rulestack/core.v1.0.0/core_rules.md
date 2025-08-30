# Core RuleStack Rules v1.0.0

This file contains the baseline rules that apply to all RuleStack projects.

## Rule Management

### Adding New Rules
When a user requests to "add a rule" or "create a rule":

1. **List Available Rule Packages**: Display all installed rule packages in .rulestack/ EXCEPT core.v1.0.0
2. **Ask for Target Package**: "Which package would you like to add this rule to?"
3. **Default to Project Rules**: If no package is specified, create/use .rulestack/project/ directory
4. **Rule File Creation**: Create appropriately named .md files with clear structure

**Example Workflow**:
```
User: "Add a rule about error handling"

Response: "I'll help you add a rule about error handling. 

Available rule packages:
- security-rules (v2.1.0)
- company-standards (v1.5.0)
- project (project-specific rules)

Which package should contain this rule? [default: project]"
```

### Project Rules Structure
- **Location**: `.rulestack/project/`
- **Purpose**: Project-specific rules that don't belong in shared packages
- **Auto-creation**: Create directory automatically when needed
- **File naming**: Use descriptive names like `error_handling.md`, `api_conventions.md`

### Rule Package Guidelines
- **Core rules** (core.v1.0.0): NEVER modify - system managed
- **Installed packages**: Add rules only with user confirmation
- **Project rules**: Default location for new project-specific rules
- **Rule organization**: Group related rules in appropriate packages

## Code Quality Rules

### Defensive Programming
- Always validate inputs and handle edge cases
- Use explicit error handling rather than silent failures
- Write clear, self-documenting code with meaningful variable names
- Include appropriate logging for debugging and monitoring

### Security Rules
- Never commit secrets, API keys, or sensitive data to repositories
- Validate and sanitize all user inputs
- Use secure coding practices appropriate for the technology stack
- Follow principle of least privilege for permissions and access

### Documentation Rules
- Document all public APIs and interfaces
- Include usage examples in code comments where helpful
- Keep README files up to date with current functionality
- Document any non-obvious business logic or algorithms

## RuleStack-Specific Rules

### Package Management
- Always run 'rfh init' before using other RuleStack commands
- Use semantic versioning for all packages
- Include clear descriptions in package manifests
- Test packages thoroughly before publishing

### Rule Development
- Write rules that are clear and actionable
- Provide examples in rule documentation
- Test rules against real-world scenarios
- Keep rules focused and single-purpose

## Integration Rules

### Claude Code Integration
- Use descriptive commit messages
- Break down large tasks into smaller, manageable steps
- Provide context when asking for code modifications
- Review generated code for correctness and style

### Version Control
- Make atomic commits with clear purposes
- Use meaningful branch names
- Keep commit history clean and readable
- Tag releases appropriately

---

*These core rules are maintained by the RuleStack system and should not be modified directly.*
