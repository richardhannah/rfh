# No Backward Compatibility Rule

## Context
RuleStack (rfh) is an unreleased project with no current users in production.

## Rule
**DO NOT** maintain backward compatibility in any part of the codebase.

## Guidelines
- Remove deprecated code immediately instead of maintaining it
- Simplify data structures without migration paths
- Change APIs, configs, and file formats freely when improvements are identified
- Don't write migration code for schema or config changes
- Remove "legacy" or "deprecated" code paths immediately

## Benefits
- Cleaner, simpler codebase
- Faster development velocity
- Reduced technical debt
- Easier to understand and maintain

## Examples
- When changing config file structure, update it directly without migration logic
- When modifying database schemas, change them without migration scripts
- When improving APIs, replace them entirely rather than versioning

## Note
This rule should be revisited once the project has active users and needs stability guarantees.